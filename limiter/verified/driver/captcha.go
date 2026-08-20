package driver

import (
	"image/color"

	"github.com/mojocn/base64Captcha"
	"github.com/thinkgos/proc-extra/limiter/verified"
)

type CaptchaDriver struct {
	alphaDigit verified.ChallengeProvider
	digit      verified.ChallengeProvider
	math       verified.ChallengeProvider
}

func NewCaptchaDriver() *CaptchaDriver {
	alphaDigitDriver := base64Captcha.NewDriverString(80, 240, 2, 2, 4, "234567890abcdefghjkmnpqrstuvwxyz",
		&color.RGBA{240, 240, 246, 246}, nil, []string{"wqy-microhei.ttc"}).
		ConvertFonts()
	digitDriver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	mathDriver := base64Captcha.NewDriverMath(80, 240, 2, 2,
		&color.RGBA{240, 240, 246, 246}, nil, []string{"wqy-microhei.ttc"}).
		ConvertFonts()

	return &CaptchaDriver{
		alphaDigit: NewCaptchaChallenge(alphaDigitDriver),
		digit:      NewCaptchaChallenge(digitDriver),
		math:       NewCaptchaChallenge(mathDriver),
	}
}

func (v *CaptchaDriver) Driver(dname string) verified.ChallengeProvider {
	switch dname {
	case "AlphaDigit":
		return v.alphaDigit
	case "Digit":
		return v.digit
	case "Math":
		return v.math
	default:
		return new(verified.UnsupportedChallengeProvider)
	}
}
