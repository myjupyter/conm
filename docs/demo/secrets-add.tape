Output secrets-add.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `D="$XDG_CONFIG_HOME/conm/default" && mkdir -p "$D" "$XDG_DATA_HOME/conm/default" && printf "%s\n" "[[postgres]]" "host = 'localhost'" "port = 15432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'orders'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'orders-local'" > "$D/postgres.toml" && printf "%s\n" "workspace_id = '00000000-0000-4000-8000-00000000c0de'" > "$XDG_DATA_HOME/conm/default/secret.toml" && clear`
Enter
Sleep 800ms
Show

Type "conm"
Sleep 300ms
Enter
Sleep 1500ms
Type "s"
Sleep 1500ms
Type "a"
Sleep 1000ms
Type "i"
Sleep 120ms
Type "orders-rw"
Sleep 150ms
Enter
Sleep 300ms
Type "i"
Sleep 120ms
Type "orders"
Sleep 150ms
Enter
Sleep 300ms
Type "i"
Sleep 120ms
Type "demo-secret"
Sleep 150ms
Enter
Sleep 300ms
Type "i"
Sleep 120ms
Type "Orders read-write user"
Sleep 150ms
Enter
Sleep 300ms
Sleep 600ms
Enter
Sleep 2500ms
