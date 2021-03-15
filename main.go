package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type M map[string]interface{}

//Renderer merupakan struct untuk render html
type Renderer struct {
	Templete *template.Template
	Debug    bool
	Location string
}

//NewRenderer mempermudah inisialisasi object renderer
func NewRenderer(Location string, Debug bool) *Renderer {
	TampR := new(Renderer)
	TampR.Location = Location
	TampR.Debug = Debug

	TampR.ReloadTemplate()
	return TampR
}

//ReloadTemplate berfungsi untuk mencari lokasi templete
func (R *Renderer) ReloadTemplate() {
	R.Templete = template.Must(template.ParseGlob(R.Location))
}

//Render adalah method untuk menampilkan atau render html di website
func (R *Renderer) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	if R.Debug {
		R.ReloadTemplate()
	}
	return R.Templete.ExecuteTemplate(w, name, data)
}

func MakeLogEntry(ctx echo.Context) *log.Entry {
	if ctx == nil {
		return log.WithFields(log.Fields{"at": time.Now().Format("2006-01-02 15:04:05")})
	}
	return log.WithFields(log.Fields{
		"at":     time.Now().Format("2006-01-02 15:04:05"),
		"method": ctx.Request().Method,
		"uri":    ctx.Request().URL.String(),
		"ip":     ctx.Request().RemoteAddr,
	})
}

func MiddleWareLogging(ctx echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		MakeLogEntry(c).Info("Incoming Request")
		return ctx(c)
	}
}

func ErrHandler(err error, ctx echo.Context) {
	report, ok := err.(*echo.HTTPError)
	if !ok {
		report.Message = fmt.Sprintf("http error %d - %v", report.Code, report.Message)
	} else {
		report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	MakeLogEntry(ctx).Error(report.Message)
	ctx.HTML(report.Code, report.Message.(string))
}

type CustomValidator struct {
	Validator *validator.Validate
}

//Validate merupakan method yang berfungsi untuk mengembalikan nilai dari validator struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}

//User merupakan struct yang berisi value yang nantinya diisi melalui json
type User struct {
	//catatan penggunaan nama variabel harus dimulai dengan kapital
	//jika tidak maka variabel tidak akan berpengaruh apa2
	Name  string `json:"name" form:"name" query:"name" validate:"required"`
	Email string `json:"email" form:"email" query:"email" validate:"required,email"`
	Age   int    `json:"age" form:"age" query:"age" validate:"gte=0,lte=80"`
}

func main() {
	r := echo.New()

	/*
		//Penggunaan middleware sebagai logger
		r.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			//Format di ganti sesuai dengan keinginan
			//sebisa mungkin tarus paling atas
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}))
	*/

	r.Use(MiddleWareLogging)
	r.HTTPErrorHandler = ErrHandler

	r.GET("/index", func(ctx echo.Context) error {
		data := "Hello from /index"
		return ctx.String(http.StatusOK, data)
	})

	r.GET("/HTML", func(ctx echo.Context) error {
		data := "Hello from /HTML"
		return ctx.HTML(http.StatusOK, data)
	})

	r.GET("/redirect", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/index")
	})

	r.GET("/json", func(ctx echo.Context) error {
		data := M{"Message": "Hello", "Counter": 2}
		return ctx.JSON(http.StatusOK, data)
	})

	//dimana any bisa berupa GET POST maupun PUT
	//jadi ini bisa digunakan dengan json xml form data maupun query string (url)
	r.Any("/user", func(ctx echo.Context) (err error) {
		u := new(User)
		if err = ctx.Bind(u); err != nil {
			return
		}
		return ctx.JSON(http.StatusOK, u)
	})

	//assign struct yang sebelumnya telah dibuat ingat gunakan
	//di file yang sama dengan pemagngilan packagenya kalo tidak bisa terjadi error
	r.Validator = &CustomValidator{Validator: validator.New()}

	//buat renderingnya
	//selamat berhasil di running dengan menggunakan validate
	r.Any("/Validate", func(ctx echo.Context) error {
		TampU := new(User)

		if err := ctx.Bind(TampU); err != nil {
			return err
		}

		if err := ctx.Validate(TampU); err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, true)
	})

	//membuat error handler lebih manusiawi dengan echo
	//jadi akan menampilkan field apa yang error tetapi masih terhubung dengan http
	/*
		r.HTTPErrorHandler = func(err error, ctx echo.Context) {
			report, ok := err.(*echo.HTTPError)
			if !ok {
				report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			ctx.Logger().Error(report)
			ctx.JSON(report.Code, report)
		}
	*/

	//membuat error lebih manusiawi lagi sehingga lebih mudah dipahami
	r.HTTPErrorHandler = func(err error, ctx echo.Context) {
		report, ok := err.(*echo.HTTPError)
		if !ok {
			report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if catedObject, ok := err.(validator.ValidationErrors); ok {
			for _, err := range catedObject {
				switch err.Tag() {
				case "required":
					report.Message = fmt.Sprintf("%s is required", err.Field())
				case "email":
					report.Message = fmt.Sprintf("%s is not valid email", err.Field())
				case "gte":
					report.Message = fmt.Sprintf("value must greater than %s", err.Param())
				case "lte":
					report.Message = fmt.Sprintf("Value must less than %s", err.Param())
				}
				break
			}
		}
		ctx.Logger().Error(report)
		ctx.JSON(report.Code, report)

		//untuk menamilkan custom page pada error
		/*
			//misal error 500 maka akan memanggil 500.html
			errPage := fmt.Sprintf("%d.html", report.code)
			//catatan File(lokasi dimana letak dari html)
			if err := ctx.File(errPage); err != nil{
				ctx.html(report.Code, "ERRORRRR")
			}
		*/
	}

	//panggil method newTemplete
	//semua bentuk html
	r.Renderer = NewRenderer("./*.html", true)

	r.GET("/Render", func(ctx echo.Context) error {
		data := M{"message": "Hello World!"}
		//nama file dan lokasi
		return ctx.Render(http.StatusOK, "render.html", data)
	})
	//catatan untuk rendering partial gunakan html/http karena
	//echo tidak bisa melakukannya

	//ayo coba middleware
	r.GET("/middle", func(ctx echo.Context) error {
		fmt.Println("aye!!!")
		return ctx.JSON(http.StatusOK, true)
	})

	/*
		r.Use(middleware.MiddleOne)
		r.Use(middleware.MiddleTwo)
		//cara memanggil 3rd party yaitu diberi tambawah WrapMiddleware
		r.Use(echo.WrapMiddleware(middleware.MiddleSomething))
	*/
	//cara menggunakan logger dari echo
	/*
		r.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}))
		//selamat logger berhasil
		r.GET("/logger", func(c echo.Context) (err error) {
			return c.JSON(http.StatusOK, true)
		})
	*/
	lock := make(chan error)
	go func(lock chan error) {
		lock <- r.Start(":9000")
	}(lock)
	time.Sleep(1 * time.Millisecond)
	MakeLogEntry(nil).Warning("Apps started without ssl/tls enabled")

	err := <-lock
	if err != nil {
		MakeLogEntry(nil).Panic("Failed to start apps")
	}

	//fmt.Println("Server start at 9000")
	//r.Start(":9000")
}
