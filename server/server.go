package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	minidns "github.com/mamad-nik/mini-dns"
)

type req struct {
	Body string `json:"body"`
}

type res struct {
	Body string `json:"body"`
}

func Serve(reqchan chan minidns.Request, multiChan chan minidns.MultiRequest) {
	router := gin.Default()

	router.LoadHTMLGlob("../../templates/*")

	singleHandler := func(ctx *gin.Context, reqtype string) {
		var r req

		if err := ctx.ShouldBindJSON(&r); err != nil {
			log.Println("error: ", err)
			return
		}
		log.Println("domain: ", r.Body)
		ipChan := make(chan string)
		errChan := make(chan error)

		reqchan <- minidns.Request{
			ReqType:  reqtype,
			Requset:  r.Body,
			Response: ipChan,
			Err:      errChan,
		}

		select {
		case ip := <-ipChan:
			ctx.JSON(http.StatusOK, res{
				Body: ip,
			})
		case err := <-errChan:
			log.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	multiHandler := func(ctx *gin.Context, reqtype string) {
		var r req

		if err := ctx.ShouldBindJSON(&r); err != nil {
			log.Println("error: ", err)
			return
		}
		log.Println("domain: ", r.Body)
		ipChan := make(chan map[string]string)
		errChan := make(chan error)

		multiChan <- minidns.MultiRequest{
			ReqType:  reqtype,
			Requset:  r.Body,
			Response: ipChan,
			Err:      errChan,
		}

		select {
		case ip := <-ipChan:
			ctx.JSON(http.StatusOK, ip)
		case err := <-errChan:
			log.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

	}

	router.GET("/favicon.ico", func(ctx *gin.Context) {
		ctx.File("../../images/gsg.webp")
	})

	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"new": true,
		})
	})

	router.POST("/dns", func(ctx *gin.Context) {
		singleHandler(ctx, "ip")
	})

	router.POST("/IP", func(ctx *gin.Context) {
		singleHandler(ctx, "domain")

	})

	router.POST("/subdomains", func(ctx *gin.Context) {
		multiHandler(ctx, "sub")
	})

	router.GET("/all", func(ctx *gin.Context) {

		ipChan := make(chan map[string]string)
		errChan := make(chan error)

		multiChan <- minidns.MultiRequest{
			ReqType:  "all",
			Requset:  "",
			Response: ipChan,
			Err:      errChan,
		}

		select {
		case ip := <-ipChan:
			ctx.JSON(http.StatusOK, ip)
		case err := <-errChan:
			log.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

	})

	s := &http.Server{
		Addr:    ":3535",
		Handler: router,
	}

	s.ListenAndServe()
}
