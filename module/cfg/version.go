package cfg

// Создание объекта и возвращение интерфейса.
func newVersion(parent *impl) Version {
	var ret = &version{parent: parent}
	return ret
}

// Major Основной номер версии.
func (ver *version) Major() uint64 { return ver.parent.main.Version.Major() }

// Minor Дополнительный номер версии.
func (ver *version) Minor() uint64 { return ver.parent.main.Version.Minor() }

// Patch Номер исправления версии.
func (ver *version) Patch() uint64 { return ver.parent.main.Version.Patch() }

// Prerelease Предварительная версия.
func (ver *version) Prerelease() string { return ver.parent.main.Version.Prerelease() }

// Metadata Метаданные версии.
func (ver *version) Metadata() string { return ver.parent.main.Version.Metadata() }

// String Представление версии в виде строки.
func (ver *version) String() string { return ver.parent.main.Version.String() }

// MarshalJSON Реализация интерфейса json.Marshaler.
func (ver *version) MarshalJSON() ([]byte, error) { return ver.parent.main.Version.MarshalJSON() }

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (ver *version) UnmarshalJSON(b []byte) error { return ver.parent.main.Version.UnmarshalJSON(b) }
