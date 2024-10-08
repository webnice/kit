package tpl

/*

Синтаксис шаблона сообщения логирования.

Пример шаблона:

${timestamp:Europe/Moscow:02.01.2006 15:04:05.000000} (${level:S:1}:${level:d:1}): ${message} {${package}/${shortfile}:${line}, функция: ${function}()}


${timestamp:TZ:FORMAT}                     - Поле: Message.Timestamp.
                                             Время записи.
                                             После первого двоеточия указывается зона времени, если не указана,
                                             используется UTC зона времени.
                                             После второго двоеточия указывается формат времени, синтаксис
                                             time.Format.layout.
                                             Пример: ${timestamp:Europe/Moscow:2006-01-02T15:04:05.000Z07:00}

${level:s:1}, ${level:d:1}                 - Поле: Message.Level.
                                             Уровень логирования сообщения.
                                             После первого двоеточия указывается тип, после второго двоеточия
                                             указывается длинна.
                                             Тип s или S - Строковое представления уровня логирования.
                                             S в верхнем регистре, выведет название в верхнем регистре.
                                             Длинна, для строки, указывает длину строки в символах, если строка меньше,
                                             тогда строка дополняется пробелами в конце.
                                             d - Числовое представление уровня логирования.
                                             Длинна для числа, указывает количество цифр, число дополняется нулями
                                             спереди.

${message:0:0}                             - Поле: Message.Pattern и Message.Argument.
                                             Шаблон и параметры шаблона, сообщения лога в виде сформированного
                                             сообщения.
                                             Число после первого двоеточия указывает минимальную длину строки.
                                             Число после второго двоеточия указывает максимальную длину строки.
                                             Если длинна сообщения меньше или больше, оно будет обрезано или дополнено
                                             пробелами в конце строки.

${keys:=:,}                                - Поле: Message.Keys.
                                             Ключи сообщения.
                                             После первого двоеточия указывается резделитель между ключём и значением.
                                             После второго двоеточия указывается разделитель между парами ключ=значение.

${stacktrace:0}                            - Поле: Message.Trace.StackTrace.
                                             Стек вызовов активного процесса, обрезанный до функции вызова.
                                             Многострочная длинная строка очерёдности вызовов.
                                             После двоеточия указывается максимальная длинна сообщения, если сообщения
                                             больше указанной длинны, оно будет обрезано в конце.

${longfile:0}                              - Поле: Message.Trace.FilenameLong.
                                             Путь и имя файла приложения из которого был совершён вызов.
                                             После двоеточия указывается максимальная длинна сообщения, если сообщения
                                             больше указанной длинны, оно будет обрезано в конце.

${shortfile:0}                             - Поле: Message.Trace.FilenameShort.
                                             Название файла из которого был совершён вызов.
                                             После двоеточия указывается максимальная длинна сообщения, если сообщения
                                             больше указанной длинны, оно будет обрезано в конце.

${function:0}                              - Поле: Message.Trace.Function.
                                             Название функции совершившей вызов.
                                             После двоеточия указывается максимальная длинна сообщения, если сообщения
                                             больше указанной длинны, оно будет обрезано в конце.

${line}                                    - Поле: Message.Trace.Line.
                                             Номер строки файла из которого был совершён вызов.

${package:0}                               - Поле: Message.Trace.Package.
                                             Название пакета.
                                             После двоеточия указывается максимальная длинна сообщения, если сообщения
                                             больше указанной длинны, оно будет обрезано в конце.


ЦВЕТА.

${dye:all:level}                           - Цвет текста и цвет фона зависящий от уровня логирования.
${dye:text:level}                          - Цвет текста зависящий от уровня логирования.
${dye:back:level}                          - Цвет фона зависящий от уровня логирования.

${dye:text:#000000}                        - Цвет текста задаётся в формате HEX RGB и конвертируется в ближайший цвет
                                             поддерживаемый терминалом.
                                             Пример: ${color:text:#FF00EE}
${dye:back:#000000}                        - Цвет фона задаётся в формате HEX RGB и конвертируется в ближайший цвет
                                             поддерживаемый терминалом.
                                             Пример: ${color:text:#CCCCEE}

${dye:text:black:normal}                   - ANSI цвет текста. Нормальный, чёрный.
${dye:text:red:normal}                     - ANSI цвет текста. Нормальный, красный.
${dye:text:green:normal}                   - ANSI цвет текста. Нормальный, зелёный.
${dye:text:yellow:normal}                  - ANSI цвет текста. Нормальный, жёлтый.
${dye:text:blue:normal}                    - ANSI цвет текста. Нормальный, синий.
${dye:text:magenta:normal}                 - ANSI цвет текста. Нормальный, пурпурный.
${dye:text:cyan:normal}                    - ANSI цвет текста. Нормальный, бирюзовый.
${dye:text:white:normal}                   - ANSI цвет текста. Нормальный, белый.

${dye:text:black:bright}                   - ANSI цвет текста. Яркий, чёрный.
${dye:text:red:bright}                     - ANSI цвет текста. Яркий, красный.
${dye:text:green:bright}                   - ANSI цвет текста. Яркий, зелёный.
${dye:text:yellow:bright}                  - ANSI цвет текста. Яркий, жёлтый.
${dye:text:blue:bright}                    - ANSI цвет текста. Яркий, синий.
${dye:text:magenta:bright}                 - ANSI цвет текста. Яркий, пурпурный.
${dye:text:cyan:bright}                    - ANSI цвет текста. Яркий, бирюзовый.
${dye:text:white:bright}                   - ANSI цвет текста. Яркий, белый.

${dye:back:black:normal}                   - ANSI цвет фона. Нормальный, чёрный.
${dye:back:red:normal}                     - ANSI цвет фона. Нормальный, красный.
${dye:back:green:normal}                   - ANSI цвет фона. Нормальный, зелёный.
${dye:back:yellow:normal}                  - ANSI цвет фона. Нормальный, жёлтый.
${dye:back:blue:normal}                    - ANSI цвет фона. Нормальный, синий.
${dye:back:magenta:normal}                 - ANSI цвет фона. Нормальный, пурпурный.
${dye:back:cyan:normal}                    - ANSI цвет фона. Нормальный, бирюзовый.
${dye:back:white:normal}                   - ANSI цвет фона. Нормальный, белый.

${dye:back:black:bright}                   - ANSI цвет фона. Яркий, чёрный.
${dye:back:red:bright}                     - ANSI цвет фона. Яркий, красный.
${dye:back:green:bright}                   - ANSI цвет фона. Яркий, зелёный.
${dye:back:yellow:bright}                  - ANSI цвет фона. Яркий, жёлтый.
${dye:back:blue:bright}                    - ANSI цвет фона. Яркий, синий.
${dye:back:magenta:bright}                 - ANSI цвет фона. Яркий, пурпурный.
${dye:back:cyan:bright}                    - ANSI цвет фона. Яркий, бирюзовый.
${dye:back:white:bright}                   - ANSI цвет фона. Яркий, белый.


РЕЖИМЫ ВЫВОДА.

${dye:reset:all}                           - Сброс цвета текста и цвета фона на цвета по умолчанию.

${dye:set:bold}                            - Вывод последующего текста жирным.
${dye:set:faded}                           - Вывод последующего текста блёклым.
${dye:set:italic}                          - Вывод последующего текста курсивом.
${dye:set:underline}                       - Вывод последующего текста подчёркнутым один раз.
${dye:set:reverse}                         - Вывод последующего текста с инвертированным цветом текста и фона.
${dye:set:crossout}                        - Вывод последующего текста зачёркнутым.
${dye:reset:bold}                          - Сброс вывода последующего текста жирным.
${dye:reset:faded}                         - Сброс вывода последующего текста блёклым.
${dye:reset:italic}                        - Сброс вывода последующего текста курсивом.
${dye:reset:underline}                     - Сброс вывода последующего текста underline.
${dye:reset:reverse}                       - Сброс вывода последующего текста с инвертированным цветом текста и фона.
${dye:reset:crossout}                      - Сброс вывода последующего текста зачёркнутым.


ФОРМАТИРОВНИЕ.

${#:}                                      - Комментарий в шаблоне, после двоеточия можно писать любой текст.
${bp--}                                    - Удаление тега и переносов строки (символы \r\n), идущих после тега.
${bp-+:0}                                  - Вставка переноса строки (символ \n) вместо тега.
                                             После первого двоеточия указывается количество вставляемых переносов.
                                             Если параметр не указан или указан равный нулю, вставляется один перенос.
${spc--:0}                                 - Удаление тега и всех пробелов идущих после тега.
                                             После первого двоеточия указывается количество вставляемых пробелов
                                             вместо тега. Если параметр не указан или указан равный нулю, не
                                             вставляется ничего.
${--spc:0}                                 - Удаление тега и всех пробелов идущих перед тегом.
                                             После первого двоеточия указывается количество вставляемых пробелов
                                             вместо тега. Если параметр не указан или указан равный нулю, не
                                             вставляется ничего.
${-spc-:0}                                 - Удаление тега и всех пробельных символов идущих как перед тегом, так и
                                             после тега.
                                             После первого двоеточия указывается количество вставляемых пробелов
                                             вместо тега. Если параметр не указан или указан равный нулю, не
                                             вставляется ничего.


*/
