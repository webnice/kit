package main

// В этом файле указываются компоненты, из которых состоит приложение.
// Каждый компонент, используя механизм зависимостей, определяет очерёдность выполнения компонентов.
// Порядок регистрации или импорта компонентов значения не имеет.
import (
	_ "github.com/webnice/kit/v4/plugin/component/bootstrap"      // Выполняется после основных зависимостей, вспомогательный компонент.
	_ "github.com/webnice/kit/v4/plugin/component/environment"    // Работа с окружением приложения.
	_ "github.com/webnice/kit/v4/plugin/component/interrupt"      // Перехват сигналов прерывания приложения.
	_ "github.com/webnice/kit/v4/plugin/component/logger_console" // Отображение сообщений логов в консоли.
	_ "github.com/webnice/kit/v4/plugin/component/pidfile"        // Работа с PID файлом приложения.
	_ "github.com/webnice/kit/v4/plugin/component/version"        // Команда отображения версии приложения.
)
