Output workspace-shell.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `mkdir -p "$XDG_CONFIG_HOME/conm/default" && printf "%s\n" "[[postgres]]" "host = 'localhost'" "port = 15432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'orders'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'orders-local'" > "$XDG_CONFIG_HOME/conm/default/postgres.toml" && clear`
Enter
Sleep 500ms
Show

Type "conm postgres list"
Sleep 300ms
Enter
Sleep 1000ms
Type "conm workspace add staging"
Sleep 300ms
Enter
Sleep 1000ms
Type "conm -w staging postgres list"
Sleep 300ms
Enter
Sleep 1200ms
Type "conm workspace use staging"
Sleep 300ms
Enter
Sleep 1000ms
Type "conm postgres list"
Sleep 300ms
Enter
Sleep 1200ms
Type "conm workspace use default"
Sleep 300ms
Enter
Sleep 1000ms
Type "conm postgres list"
Sleep 300ms
Enter
Sleep 2000ms
