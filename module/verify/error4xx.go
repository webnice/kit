package verify

// E4xx HTTP error code 400-499
func E4xx() Interface {
	var err = new(Response)
	err.Error.Code = 4
	err.Error.Message = `Data is incorrect`
	return err
}
