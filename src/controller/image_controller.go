package controller

import (
	"argos/src/models/image"
	"github.com/gin-gonic/gin"
	"net/http"
)

type imageController struct {
	service image.Service
}

// @BasePath /argos

// RegisterImage @Description get data by Image rul
// @Accept  json
// @Produce  json
// @Param   image_url     query    string     true    "https://google.com/img.jpeg"
// @Success 200 {string} json	"ok"
// @Failure 400 {object} rest_error.RestErr "We need image_url!!"
// @Failure 404 {object} rest_error.RestErr "Can not find image_url"
// @Router /register_image [get]
func (ic imageController) RegisterImage(c *gin.Context) {
	args := struct {
		ImageUrl string `form:"image_url" binding:"required"`
	}{}

	if err := c.BindQuery(&args); err != nil {
		c.JSON(400, gin.H{
			"msg": err.Error()})
		return
	}

	err := ic.service.DownloadImage(args.ImageUrl)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Message()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "downloaded"})
}

// @BasePath /argos

// GetImage @Description get data by Image rul
// @Accept  json
// @Produce  json
// @Param   image_url     query    string     true    "https://google.com/img.jpeg"
// @Success 200 {string} data	"ok"
// @Failure 400 {object} rest_error.RestErr "We need image_url!!"
// @Failure 404 {object} rest_error.RestErr "Can not find image_url"
// @Router /get_image [get]
func (ic imageController) GetImage(c *gin.Context) {
	args := struct {
		ImageUrl string `form:"image_url" binding:"required"`
	}{}

	if err := c.BindQuery(&args); err != nil {
		c.JSON(400, gin.H{
			"msg": err.Error()})
		return
	}

	image, err := ic.service.GetImage(args.ImageUrl)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Message()})
		return
	}
	c.Data(http.StatusOK, "image/jpeg", image)
}

type SiteMetaInterfaces interface {
	GetImage(c *gin.Context)
	RegisterImage(c *gin.Context)
	Route(prefix *gin.RouterGroup)
}

func NewHandler(service image.Service) SiteMetaInterfaces {
	return &imageController{service: service}
}

func (ic *imageController) Route(prefix *gin.RouterGroup) {
	prefix.GET(
		"/get_image",
		ic.GetImage,
	)
	prefix.GET(
		"/register_image",
		ic.RegisterImage,
	)
}
