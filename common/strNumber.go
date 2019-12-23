package common

import (
	"github.com/shopspring/decimal"
)

func BcAdd(a string, b string, precision int32) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "0", err
	}

	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "0", err
	}

	aa = aa.Add(bb)
	return aa.StringFixedBank(precision), nil
}

func BcSub(a string, b string, precision int32) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "0", err
	}

	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "0", err
	}

	aa = aa.Sub(bb)
	return aa.StringFixedBank(precision), nil
}

func BcAbsSub(a string, b string, precision int32) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "0", err
	}

	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "0", err
	}

	aa = aa.Sub(bb)
	aa = aa.Abs()
	return aa.StringFixedBank(precision), nil
}

func BcDiv(a string, b string, precision int32) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "0", err
	}

	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "0", err
	}

	aa = aa.Div(bb)
	return aa.StringFixedBank(precision), nil
}

func BcMul(a string, b string, precision int32) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "0", err
	}

	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "0", err
	}

	aa = aa.Mul(bb)
	return aa.StringFixedBank(precision), nil
}

func BcPow(a string, b string, precision int32) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "0", err
	}

	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "0", err
	}

	aa = aa.Pow(bb)
	return aa.StringFixedBank(precision), nil
}

func BcMod(a string, b string, precision int32) (string, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return "0", err
	}

	bb, err := decimal.NewFromString(b)
	if err != nil {
		return "0", err
	}

	aa = aa.Mod(bb)
	return aa.StringFixedBank(precision), nil
}

func BcCmp(a string, b string) (int, error) {
	aa, err := decimal.NewFromString(a)
	if err != nil {
		return 0, err
	}

	bb, err := decimal.NewFromString(b)
	if err != nil {
		return 0, err
	}

	return aa.Cmp(bb), nil
}

var (
	tenPow18 string
)

func init() {
	var err error
	tenPow18, err = BcPow("10", "18", 0)
	if err != nil {
		panic(err)
	}
}

// 普通金额字符串转换为区块链38位金额字符串
func Money2BlockMoney(money string) string {
	if money == "" {
		return "0"
	}

	a, err := BcMul(money, tenPow18, 0)
	if err != nil {
		return "0"
	}

	return a
}

// 区块链38位金额字符串转普通字符串
func BlockMoney2Money(blockMoney string, precision int32) string {
	if blockMoney == "" {
		return "0"
	}

	a, err := BcDiv(blockMoney, tenPow18, precision)
	if err != nil {
		return "0"
	}

	return a
}
