package main

import (
	"errors"
	"fmt"
)

//第一周作业
//完成 切片的删除方法(通过下标来删除切片元素) 通过泛型支持 多种类型

func main() {
	s1 := []int{1, 2, 3, 4}
	s2, e1 := DelSlice[int](s1, 3)

	if e1 != nil {
		fmt.Printf("error %v", e1)
	} else {
		fmt.Printf(" %v", s2)
	}

	s3 := []string{"a", "b", "c", "d"}
	s4, e2 := DelSlice[string](s3, 1)
	if e2 != nil {
		fmt.Printf("error %v", e2)
	} else {
		fmt.Printf(" %v", s4)
	}

}

type indexed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~uintptr | ~float32 | ~float64 | ~string
}

func DelSlice[T indexed](s []T, i int) ([]T, error) {
	n := len(s)

	if n == 0 {
		return []T{}, errors.New("slice is empty ")
	}

	if i < 0 || i >= len(s) {
		return []T{}, errors.New("index is illegality! ")
	}
	if i == 0 {
		return s[1:], nil
	}
	if i == n-1 {
		return s[:n-1], nil
	}
	var rs []T
	s1 := s[:i]
	s2 := s[i+1:]
	rs = append(s1, s2...)

	return rs, nil
}
