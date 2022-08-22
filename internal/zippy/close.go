package zippy

func (z ZipReader) Close() {
	_ = z.reader.Close()
}
