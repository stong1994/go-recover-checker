package main

func main() {
	checker := NewChecker(nil)
	files := []string{"./data/"}
	if err := checker.ParseFiles(files); err != nil {
		panic(err)
	}
	checker.DisplayNeedRecoverList()
}
