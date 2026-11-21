Output edit-table.gif

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

Type "conm postgres"
Sleep 300ms
Enter
Sleep 1500ms
Type "e"
Sleep 1200ms
Tab
Sleep 300ms
Tab
Sleep 600ms
Type "j"
Sleep 300ms
Type "i"
Sleep 120ms
Type "Local orders database"
Sleep 150ms
Enter
Sleep 400ms
Sleep 600ms
Enter
Sleep 1500ms
Type "p"
Sleep 2000ms
Type "q"
Sleep 800ms
