package translator

import (
	"regexp"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	Validate *validator.Validate
	trans    ut.Translator
)

func InitValidator(localeInfo ...string) error {
	Validate = validator.New()
	var err error
	// 注册一个函数，获取struct tag里自定义的label作为字段名
	//Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
	//	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	//	if name == "-" {
	//		return ""
	//	}
	//	return name
	//})

	// 在这里注册自定义结构体/字段校验方法
	// 注册自定义结构体校验方法
	//validate.RegisterStructValidation(SignUpParamStructLevelValidation, TestReq{})

	// 注册自定义结构体字段校验方法
	//if err := validate.RegisterValidation("checkDate", lottery.CheckDate); err != nil {
	//	return err
	//}
	// 注册一个叫 "password" 的验证器
	if err = Validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		reLetter := regexp.MustCompile(`[A-Za-z]`)
		reDigit := regexp.MustCompile(`\d`)
		// 至少 6 位，包含字母和数字
		return len(password) >= 6 && reLetter.MatchString(password) && reDigit.MatchString(password)
	}); err != nil {
		return err
	}

	// 在这里注册自定义tag翻译
	// 注意！因为这里会使用到trans实例
	// 所以这一步注册要放到trans初始化的后面

	//if err := validate.RegisterTranslation(
	//	"checkDate",
	//	trans,
	//	registerTranslator("checkDate", "{0}必须要晚于当前日期"),
	//	translate,
	//); err != nil {
	//	return err
	//}

	// 验证器注册翻译器
	var trans ut.Translator
	locale := "en"
	if len(localeInfo) > 0 {
		locale = localeInfo[0]
	}
	switch locale {
	case "en":
		uni := ut.New(en.New())
		trans, _ = uni.GetTranslator("en")
		err = enTranslations.RegisterDefaultTranslations(Validate, trans)
	case "zh":
		uni := ut.New(zh.New())
		trans, _ = uni.GetTranslator("zh")
		err = zhTranslations.RegisterDefaultTranslations(Validate, trans)
	default:
		uni := ut.New(en.New())
		trans, _ = uni.GetTranslator("en")
		err = enTranslations.RegisterDefaultTranslations(Validate, trans)
	}
	if err != nil {
		return err
	}
	return nil
}

// ValidateStruct 统一的验证方法
func ValidateStruct(data interface{}) error {
	if err := Validate.Struct(data); err != nil {
		//for _, err := range err.(validator.ValidationErrors) {
		//	return errors.New(err.Translate(trans))
		//}
		// if _, ok := err.(validator.ValidationErrors); ok {
		// 	// 不返回每个字段的具体错误，统一返回参数错误
		// 	return errors.New(xerr.MapErrMsg(xerr.RequestParamError))
		// }
		return err
	}
	return nil
}

// Test Example
//func main() {
//	_ = InitValidator("zh")
//
//	req := RegisterRequest{
//		Email:    "abc",
//		Password: "123",
//		ConfirmPassword: "456",
//	}
//
//	if err := ValidateStruct(req); err != nil {
//		fmt.Println("校验失败:", err.Error())
//	}
//}

// registerTranslator 为自定义字段添加翻译功能
func registerTranslator(tag string, msg string) validator.RegisterTranslationsFunc {
	return func(trans ut.Translator) error {
		if err := trans.Add(tag, msg, false); err != nil {
			return err
		}
		return nil
	}
}

// translate 自定义字段的翻译方法
func translate(trans ut.Translator, fe validator.FieldError) string {
	msg, err := trans.T(fe.Tag(), fe.Field())
	if err != nil {
		panic(any(fe.(error).Error()))
	}
	return msg
}
