package image

import (
	"argos/config"
	"argos/src/models/worker_result"
	"argos/src/repository"
	"argos/src/utils"
	"argos/src/utils/image_processing"
	"argos/src/utils/lfu"
	"argos/src/utils/rest_error"
	"argos/src/utils/workerpool"
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/h2non/bimg"
	log "github.com/sirupsen/logrus"
	"math"
	"time"
)

var (
	ErrNon200        = errors.New("received non 200 response code")
	ErrImageNotFound = errors.New("image not found")
	requestCounter   = 0
)

type Service interface {
	Translate(maxWidth, maxHeight int, format bimg.ImageType) (buffer []byte, err error)
	CacheImage(url string, file []byte, maxWidth, maxHeight int, format bimg.ImageType)
	CacheUrl(url string)
	GetInvalidUrl(url string) bool
	GetImage(url string) (buffer []byte, err rest_error.RestErr)
	DownloadImage(url string) rest_error.RestErr
}
type image struct {
	cache      config.LocalCache
	lfuCache   *lfu.Cache
	repository repository.ImageRepository
	wp         workerpool.WorkerPool
	*bimg.Image
	MaxWidth, MaxHeight int
	entry               *log.Entry
}

func (i image) GetInvalidUrl(url string) bool {
	ctx := context.Background()
	_, err := i.repository.GetInvalidUrl(ctx, url)
	if err == redis.Nil {
		return false
	}
	return true
}

func (i image) CacheUrl(url string) {
	i.repository.CacheInvalidUrl(url)
}

func (i image) DownloadImage(url string) (error rest_error.RestErr) {
	//we Evict lfuCache every 50 request and  reset requestCounter to 0
	if requestCounter == 50 {
		i.lfuCache.Evict(1)
		requestCounter = 0
	}
	requestCounter += 1

	//get image form lfu cache
	cache := i.lfuCache.Get(url)
	if cache != nil {
		i.entry.Infof("This  %s is already in lfu cache\n", url)
		return
	}

	//check url is invalid or not
	b := i.GetInvalidUrl(url)
	if b {
		error = rest_error.NewNotFoundError("this url is not valid and couldn't fine any image from this url")
		return
	}

	ctx := context.Background() //todo add timeout
	_, err := i.repository.GetImage(ctx, url)

	if err == redis.Nil {
		//create name spaces for local cache for is urk in download progress or not
		_, ok := i.cache.Get(url)
		if ok {
			i.entry.Infof("This  %s in download proggress\n", url)
			return nil
		}

		//set url to cache for check is url in download
		//progress and after 30 second is going to remove url form local cache
		i.cache.SetWithTTL(url, 1, 0, 30*time.Second)

		//add download and save to redis task with worker
		//create chan for return date or error from download
		result := make(chan worker_result.Result)
		defer close(result)

		//add download and save to redis task with worker
		i.wp.AddTask(func() {
			response, err := utils.DownloadFile(url)
			if err != nil {
				i.entry.Errorf("this  %s is not valid", url)
				go i.CacheUrl(url)
				result <- worker_result.Result{
					Rrr:    err,
					Status: 404,
					Value:  nil,
				}
				return
			}
			//check the image format
			//if image format is not valid we cache the image on redis to not in progress again
			formate, errF := image_processing.GuessImageFormat(response)
			if errF != nil {
				i.entry.Errorf("this  %s is not image, format of file is %s", url, formate)
				go i.CacheUrl(url)
				result <- worker_result.Result{
					Rrr:    err,
					Status: 404,
					Value:  nil,
				}
				return
			}
			//cache image data with config and image format
			//we return from result chan to return as bytes for user requested
			go i.CacheImageWithChan(url, response, i.MaxWidth, i.MaxHeight, bimg.JPEG, result)
		})
		res := <-result
		if res.Status == 404 {
			error = rest_error.NewNotFoundError("this url is not valid and couldn't fine any image from this url")
		}
		return
	}
	return nil
}

func (i image) GetImage(url string) (buffer []byte, error rest_error.RestErr) {
	//we Evict lfuCache every 50 request and  reset requestCounter to 0
	if requestCounter == 50 {
		i.lfuCache.Evict(1)
		requestCounter = 0
	}
	requestCounter += 1

	//get image form lfu cache
	cache := i.lfuCache.Get(url)
	if cache != nil {
		i.entry.Infof("This  %s is already in lfu cache\n", url)
		buffer = cache.([]byte)
		return
	}

	//check url is invalid or not
	b := i.GetInvalidUrl(url)
	if b {
		error = rest_error.NewNotFoundError("this url is not valid and couldn't fine any image from this url")
		return
	}

	ctx := context.Background()
	bufferStr, err := i.repository.GetImage(ctx, url)
	if err == redis.Nil {
		//create name spaces for local cache for is urk in download progress or not
		_, ok := i.cache.Get(url)
		if ok {
			i.entry.Infof("This  %s in download proggress\n", url)
			return nil, nil
		}
		//set url to cache for check is url in download
		//progress and after 30 second is going to remove url form local cache
		i.cache.SetWithTTL(url, 1, 0, 30*time.Second)

		//create chan for return date or error from download
		result := make(chan worker_result.Result)
		defer close(result)

		//add download and save to redis task with worker
		i.wp.AddTask(func() {
			response, err := utils.DownloadFile(url)
			if err != nil {
				go i.CacheUrl(url)
				result <- worker_result.Result{
					Rrr:    err,
					Status: 404,
					Value:  nil,
				}
				return
			}
			//check the image format
			//if image format is not valid we cache the image on redis to not in progress again
			formate, errF := image_processing.GuessImageFormat(response)
			if errF != nil {
				i.entry.Errorf("this url %s is not image, format of file is %s", url, formate)
				go i.CacheUrl(url)
				result <- worker_result.Result{
					Rrr:    err,
					Status: 404,
					Value:  nil,
				}
				return
			}
			//cache image data with config and image format
			//we return from result chan to return as bytes for user requested
			go i.CacheImageWithChan(url, response, i.MaxWidth, i.MaxHeight, bimg.JPEG, result)
		})
		res := <-result
		if res.Status == 200 {
			buffer = res.Value
		} else {
			error = rest_error.NewNotFoundError("this url is not valid and couldn't fine any image from this url")
		}
		return
	}
	//create byte from redis
	i.entry.Infof("This  %s is already in redis cache\n", url)
	buffer = []byte(bufferStr)
	return
}

func (i image) Translate(maxWidth, maxHeight int, format bimg.ImageType) (buffer []byte, err error) {
	imageSize, _ := i.Size()
	var ratio float64
	var newSize bimg.ImageSize

	if imageSize.Width >= imageSize.Height {
		ratio = float64(maxHeight) / float64(imageSize.Height)
		if ratio >= 1.0 {
			goto JustConvert
		}
		newSize.Width = int(math.Floor(float64(imageSize.Width) * ratio))
		newSize.Height = int(math.Floor(float64(imageSize.Height) * ratio))
	} else if imageSize.Width < imageSize.Height {
		ratio = float64(maxWidth) / float64(imageSize.Width)
		if ratio >= 1.0 {
			goto JustConvert
		}
		newSize.Width = int(math.Floor(float64(imageSize.Width) * ratio))
		newSize.Height = int(math.Floor(float64(imageSize.Height) * ratio))
	}
	_, err = i.ForceResize(newSize.Width, newSize.Height)
	if err != nil {
		return
	}
JustConvert:
	buffer, err = i.Convert(format)
	if err != nil {
		return
	}
	return
}

func (i image) CacheImage(url string, file []byte, maxWidth, maxHeight int, format bimg.ImageType) {
	i.Image = NewImg(file)
	buffer, _ := i.Translate(maxWidth, maxHeight, format)
	i.lfuCache.Set(url, buffer)
	i.repository.CacheImage(url, buffer)
}
func (i image) CacheImageWithChan(url string, file []byte, maxWidth, maxHeight int, format bimg.ImageType, result chan worker_result.Result) {
	i.Image = NewImg(file)
	buffer, _ := i.Translate(maxWidth, maxHeight, format)
	i.repository.CacheImage(url, buffer)
	i.lfuCache.Set(url, buffer)
	result <- worker_result.Result{
		Rrr:    nil,
		Status: 0,
		Value:  buffer,
	}
}
func NewService(repo *repository.ImageRepository, wp workerpool.WorkerPool, maxWidth, maxHeight int, cache config.LocalCache, lf *lfu.Cache, entry *log.Entry) Service {
	return &image{
		lfuCache:   lf,
		cache:      cache,
		wp:         wp,
		repository: *repo,
		Image:      nil,
		MaxWidth:   maxWidth,
		MaxHeight:  maxHeight,
		entry:      entry,
	}
}

func NewImg(buffer []byte) *bimg.Image {
	return bimg.NewImage(buffer)
}
