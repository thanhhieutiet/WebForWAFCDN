package money

import "fmt"

// VND represents an amount in Vietnamese Dong.
type VND int64

// Zero returns a zero VND amount.
const Zero VND = 0

// Add adds two VND amounts.
func (v VND) Add(other VND) VND {
	return v + other
}

// Sub subtracts other VND from v.
func (v VND) Sub(other VND) VND {
	return v - other
}

// Mul multiplies v by an integer factor.
func (v VND) Mul(n int64) VND {
	return v * VND(n)
}

// Int64 returns the underlying int64 value.
func (v VND) Int64() int64 {
	return int64(v)
}

// Format formats the VND amount as a Vietnamese currency string, e.g. "100.000 ₫".
func (v VND) Format() string {
	n := int64(v)
	isNegative := n < 0
	if isNegative {
		n = -n
	}

	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		if isNegative {
			return "-" + str + " ₫"
		}
		return str + " ₫"
	}

	var res []byte
	rem := len(str) % 3
	if rem > 0 {
		res = append(res, str[:rem]...)
	}
	for i := rem; i < len(str); i += 3 {
		if len(res) > 0 {
			res = append(res, '.')
		}
		res = append(res, str[i:i+3]...)
	}

	if isNegative {
		return "-" + string(res) + " ₫"
	}
	return string(res) + " ₫"
}
