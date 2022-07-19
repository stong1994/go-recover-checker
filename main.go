package main

func main() {
	checker := NewChecker(nil)
	//files := []string{"./a/"}
	files := []string{"C:\\Project\\internal\\dm.bp.adapters\\service\\wework_secret_sync\\"}
	if err := checker.ParseFiles(files); err != nil {
		panic(err)
	}
	checker.DisplayNeedRecoverList()
}
