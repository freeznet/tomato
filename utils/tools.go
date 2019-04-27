package utils

import (
	"fmt"
	"github.com/freeznet/tomato/errs"
	"reflect"
	"regexp"
	"strconv"

	"github.com/freeznet/tomato/types"
)

// HasResults Find() 返回数据中是否有结果
func HasResults(response types.M) bool {
	if response == nil ||
		response["results"] == nil ||
		A(response["results"]) == nil ||
		len(A(response["results"])) == 0 {
		return false
	}
	return true
}

// IsEmail ...
func IsEmail(email string) bool {
	b, _ := regexp.MatchString("^.+@.+$", email)
	return b
}

// DeepCopy 简易版的内存复制
func DeepCopy(i interface{}) interface{} {
	return Copy(i)
}

// CopyMap 复制 map
func CopyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	d := map[string]interface{}{}
	for k, v := range m {
		d[k] = DeepCopy(v)
	}
	return d
}

// CopySlice 复制 slice
func CopySlice(s []interface{}) []interface{} {
	if s == nil {
		return nil
	}
	d := []interface{}{}
	for _, v := range s {
		d = append(d, DeepCopy(v))
	}
	return d
}

// CopyMapM 复制 map
func CopyMapM(m types.M) types.M {
	if m == nil {
		return nil
	}
	d := types.M{}
	for k, v := range m {
		d[k] = DeepCopy(v)
	}
	return d
}

// CopySliceS 复制 slice
func CopySliceS(s types.S) types.S {
	if s == nil {
		return nil
	}
	d := types.S{}
	for _, v := range s {
		d = append(d, DeepCopy(v))
	}
	return d
}

// CompareArray 比较两个数组是否相等，忽略数组顺序
func CompareArray(i1, i2 interface{}) bool {
	if i1 == nil && i2 == nil {
		return true
	}
	if v1 := A(i1); v1 != nil {
		if v2 := A(i2); v2 != nil {
			// TODO 去重
			if len(v1) != len(v2) {
				return false
			}

			for _, a := range v1 {
				match := false
				for _, b := range v2 {
					if reflect.DeepEqual(a, b) {
						match = true
						break
					}
				}
				if match == false {
					return false
				}
			}
			return true
		}
		return false
	}
	return false
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

/*
* Throws an exception if the given lat-long is out of bounds.
*/
func ValidatePolygonPoint(point1, point2 interface{}) error {
	var latitude float64
	var longitude float64
	if _, ok := point1.(float64); ok {
		latitude = point1.(float64)
	} else if i, ok := point1.(int); ok {
		if i < -90 {
			return errs.E(errs.InvalidJSON, fmt.Sprintf("GeoPoint latitude out of bounds: ' %d ' < -90.0.", i))
		}
		if i > 90 {
			return errs.E(errs.InvalidJSON, fmt.Sprintf("GeoPoint latitude out of bounds: ' %d ' > -90.0.", i))
		}
	}else if s, ok := point1.(string); ok {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errs.E(errs.InvalidJSON, "GeoPoint latitude and longitude must be valid numbers")
		} else {
			latitude = f
		}
	} else {
		return errs.E(errs.InvalidJSON, "GeoPoint latitude and longitude must be valid numbers")
	}

	if _, ok := point2.(float64); ok {
		longitude = point2.(float64)
	} else if i, ok := point2.(int); ok {
		if i < -180 {
			return errs.E(errs.InvalidJSON, fmt.Sprintf("GeoPoint longitude out of bounds: ' %d ' < 180.0.", i))
		}
		if i > 180 {
			return errs.E(errs.InvalidJSON, fmt.Sprintf("GeoPoint longitude out of bounds: ' %d ' > 180.0.", i))
		}
	}else if s, ok := point2.(string); ok {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errs.E(errs.InvalidJSON, "GeoPoint latitude and longitude must be valid numbers")
		} else {
			longitude = f
		}
	} else {
		return errs.E(errs.InvalidJSON, "GeoPoint latitude and longitude must be valid numbers")
	}

	if latitude < -90.0 {
		return errs.E(errs.InvalidJSON, fmt.Sprintf("GeoPoint latitude out of bounds: ' %d ' < -90.0.", latitude))
	}
	if latitude > 90.0 {
		return errs.E(errs.InvalidJSON, fmt.Sprintf("GeoPoint latitude out of bounds: ' %d ' > -90.0.", latitude))
	}
	if longitude < -180.0 {
		return errs.E(errs.InvalidJSON, fmt.Sprintf("GeoPoint longitude out of bounds: ' %d ' < 180.0.", longitude))
	}
	if longitude > 180.0 {
		return errs.E(errs.InvalidJSON, fmt.Sprintf("GeoPoint longitude out of bounds: ' %d ' > 180.0.", longitude))
	}
	return nil
}