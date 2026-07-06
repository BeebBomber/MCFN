# MCFN — My Configuration for NixOS

##### From empty console to complete Flakes ecosystem in minutes.

MCFN (My Configuration for NixOS) — это полнофункциональный CLI-инструмент для автоматизации настройки и развертывания NixOS. Проект объединяет возможности Flakes, Home Manager, Stylix и Sops-nix, предоставляя интерфейс как для начинающих пользователей, так и для профи.

---

## Способы запуска

### Прямой запуск (рекомендуется)
Запуск актуальной версии напрямую из репозитория без предварительной установки:
```bash
nix run github:BeebBomber/mcfn
```

### Экспериментальный метод (Bootstrap)
Для быстрой развертки структуры проекта одной командой (используется ветка experimental):
```bash
curl -sSL https://raw.githubusercontent.com/BeebBomber/mcfn/experimental/bootstrap.sh | bash
```

---

## Ключевые возможности

### Глобальные функции
* Multi-language: Автоматическое определение RU/EN языка интерфейса и генерируемых файлов конфигурации.
* Wizard and Pro Modes: Выбор между автоматизированным мастером и полным ручным контролем над каждым параметром.
* Validation: Обязательная проверка кода через nix-instantiate и nix flake check перед применением изменений.
* Safe-Apply: Автоматическое резервное копирование текущей конфигурации в /etc/nixos.bak.

### Технический стек
* Remote Deploy: Установка и обновление NixOS на удаленных хостах по SSH.
* Community Presets: Готовые наборы настроек для различных сценариев (Gaming, Development, Minimal).
* ISO Builder: Генерация собственного установочного образа на основе текущего конфига.
* VM Test: Сборка и запуск конфигурации в виртуальной машине QEMU для тестирования.
* Secret Management: Интеграция sops-nix для управления зашифрованными паролями.
* Theming: Единая система темизации через Stylix.

---

## Архитектура создаваемой конфигурации

Инструмент генерирует модульную структуру в директории /etc/nixos/:

```text
/etc/nixos/
├── flake.nix             # Управление зависимостями и входами
├── configuration.nix     # Глобальные системные настройки
├── home.nix              # Пользовательские настройки (Home Manager)
├── hardware-config.nix   # Конфигурация оборудования (автоопределение)
└── modules/              # Дополнительные модули и секреты
```

---

## Инструкция по работе с GitHub Cloud

Для синхронизации конфигурации MCFN использует Personal Access Tokens (PAT).
1. Перейдите в Settings -> Developer settings -> Tokens (classic).
2. Создайте новый токен с разрешением (scope) "repo".
3. Скопируйте токен и введите его при запросе в MCFN.

---

## Разработка

Для подготовки среды разработки используйте Nix:

```bash
git clone https://github.com/username/mcfn.git
cd mcfn
nix develop
```

Конфигуратор MCFN, 2026.