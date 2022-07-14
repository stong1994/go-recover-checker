package main

func main() {
	checker := NewChecker(nil)
	//files := []string{"./data/"}
	files := []string{"C:\\Project\\internal\\dm.bp.adapters"}
	if err := checker.ParseFiles(files); err != nil {
		panic(err)
	}
	checker.DisplayNeedRecoverList()
}
