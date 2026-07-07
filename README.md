# MCFN — NixOS Config Generator

TUI-инструмент для генерации `configuration.nix` и `flake.nix` под вашу NixOS-систему. Запускается как пошаговый визард, не требует ручного редактирования Nix-файлов.

## Возможности

- Пошаговый TUI-визард (bubbletea)
- Генерация `configuration.nix` и опционально `flake.nix`
- **Рабочие столы:** GNOME, KDE Plasma 6, XFCE, Cinnamon, MATE, LXQt, i3, Sway, Hyprland, Niri
- **Дисплей-менеджеры:** GDM, SDDM, LightDM, greetd, ly — выбираются независимо от DE
- **Файловые системы:** ext4, btrfs (autoScrub), ZFS (autoScrub + случайный hostId)
- **Home Manager** — как nixosModule в flake
- **sops-nix** — шаблон secrets + .sops.yaml
- Несколько пользователей (основной + дополнительные)
- Предпросмотр файлов перед генерацией (`Tab` — переключение, `↑/↓` — прокрутка)
- Загрузка конфига в приватный GitHub-репозиторий (токен, создаёт репо автоматически)
- Сохранение настроек в `mcfn-config.json` для повторного использования

## Сборка

Требуется Go 1.22+.

```bash
git clone https://github.com/BeebBomber/MCFN.git
cd MCFN
go build -o mcfn ./cmd/mcfn/
```

Или через Nix:

```bash
nix build
```

## Использование

```bash
./mcfn                                          # интерактивный визард
./mcfn --hostname mypc --username alice         # предзаполнить поля
./mcfn --output /etc/nixos --no-github          # сохранить в /etc/nixos, без GitHub
./mcfn --config ./nixos-config/mcfn-config.json # повторить прошлую конфигурацию
```

### Флаги

| Флаг | Описание |
|------|----------|
| `-h`, `--help` | Справка |
| `--hostname <name>` | Предзаполнить hostname |
| `--username <name>` | Предзаполнить имя пользователя |
| `--output <dir>` | Директория для сохранения (default: `./nixos-config`) |
| `--no-github` | Пропустить шаг загрузки в GitHub |
| `--config <file>` | Загрузить настройки из `mcfn-config.json` |

## Шаги визарда

1. Hostname
2. Основной пользователь
3. Дополнительные пользователи
4. flake.nix (да/нет)
5. Загрузчик (systemd-boot, GRUB EFI, GRUB Legacy)
6. Архитектура (x86\_64, aarch64)
7. Версия NixOS (25.05, 26.05, unstable)
8. Рабочий стол
9. Дисплей-менеджер
10. Файловая система
11. NetworkManager
12. SSH
13. Home Manager *(только при flake)*
14. sops-nix *(только при flake)*
15. Дополнительные пакеты
16. Директория сохранения
17. Подтверждение + предпросмотр файлов
18. Генерация
19. Загрузка в GitHub *(опционально)*

## Применение конфига

```bash
# С flake.nix
sudo nixos-rebuild switch --flake ./nixos-config#hostname

# Без flake.nix
sudo nixos-rebuild switch -I nixos-config=./nixos-config/configuration.nix
```

## GitHub

Для загрузки конфига нужен Personal Access Token с правами `repo`:

1. GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Создать токен с разрешением **repo**
3. Вставить при запросе в визарде

MCFN создаст **приватный** репозиторий и запушит туда конфиг.

## Генерируемые файлы

```
nixos-config/
├── configuration.nix     # основной конфиг системы
├── flake.nix             # если выбрано
├── home.nix              # если выбран Home Manager
├── secrets/
│   └── secrets.yaml      # если выбран sops-nix (заглушка под шифрование)
├── .sops.yaml            # если выбран sops-nix
└── mcfn-config.json      # настройки для --config
```
