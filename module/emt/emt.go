package emt

import (
	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitTypes "github.com/webnice/kit/v4/types"
)

// New Конструктор объекта сущности пакета, возвращается интерфейс пакета.
func New() Interface {
	var emt = &impl{
		cfg: kitModuleCfg.Get(),
	}

	return emt
}

// Ссылка на менеджер логирования.
func (emt *impl) log() kitTypes.Logger { return emt.cfg.Log() }
