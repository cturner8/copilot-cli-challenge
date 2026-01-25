package install

func ExtractAsset(srcPath string, destDir string) error {
	// TODO: check file type
	return extractZip(srcPath, destDir)
}
