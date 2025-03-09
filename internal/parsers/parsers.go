package parsers

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"time"
)

type parseFunc[T any] func(s string) (T, error)

func Ptr[T any](f parseFunc[T]) func(string) (*T, error) {
	return func(s string) (*T, error) {
		v, err := f(s)
		return &v, err
	}
}

func String(s string) (string, error) {
	return s, nil
}

func Inte64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func Uint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func Uint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func Float64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func Time(layout string) func(string) (time.Time, error) {
	return func(s string) (time.Time, error) {
		return time.Parse(layout, s)
	}
}

func URL(s string) (url.URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return url.URL{}, err
	}
	return *u, nil
}

func IP(s string) (net.IP, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, errors.New("invalid IP address")
	}
	return ip, nil
}
