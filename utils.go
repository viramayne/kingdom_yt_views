package main

func ConcatinateArrays(arr1, arr2 []string) *[]string {
	for i := range arr2 {
		arr1 = append(arr1, arr2[i])
	}
	return &arr1
}
