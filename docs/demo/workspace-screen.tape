Output workspace-screen.gif

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

Type "conm"
Sleep 300ms
Enter
Sleep 1800ms
Type "w"
Sleep 1200ms
Type "a"
Sleep 800ms
Type "i"
Sleep 120ms
Type "staging"
Sleep 150ms
Enter
Sleep 300ms
Enter
Sleep 1200ms
Type "j"
Sleep 400ms
Type "e"
Sleep 800ms
Type "i"
Sleep 120ms
Type "-eu"
Sleep 150ms
Enter
Sleep 300ms
Enter
Sleep 1200ms
Enter
Sleep 2500ms
