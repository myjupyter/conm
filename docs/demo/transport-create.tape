Output transport-create.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `D="$XDG_CONFIG_HOME/conm/default" && mkdir -p "$D" && printf "%s\n" "[[postgres]]" "host = 'postgres'" "port = 5432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'orders'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'orders'" > "$D/postgres.toml" && printf "%s\n" "[[ssh]]" "host = 'localhost'" "port = 2222" "username = 'demo'" "auth = 'password'" "password = 'demo-secret'" "" "[ssh.meta]" "name = 'bastion'" > "$D/ssh.toml" && clear`
Enter
Sleep 500ms
Show

Type "conm"
Sleep 300ms
Enter
Sleep 1500ms
Type "t"
Sleep 1200ms
Type "a"
Sleep 1000ms
Type "i"
Sleep 120ms
Type "via-bastion"
Sleep 150ms
Enter
Sleep 300ms
Type "j"
Sleep 800ms
Enter
Sleep 2500ms
