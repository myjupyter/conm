Output transport-direct.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `D="$XDG_CONFIG_HOME/conm/default" && mkdir -p "$D" && printf "%s\n" "[[postgres]]" "host = 'postgres'" "port = 5432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'orders'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'orders'" > "$D/postgres.toml" && clear`
Enter
Sleep 500ms
Show

Type "conm postgres ping orders"
Sleep 300ms
Enter
Sleep 3500ms
