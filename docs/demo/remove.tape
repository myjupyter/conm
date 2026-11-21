Output remove.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `mkdir -p "$XDG_CONFIG_HOME/conm/default" && printf "%s\n" "[[postgres]]" "host = 'localhost'" "port = 15432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'orders'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'orders-local'" "" "[[postgres]]" "host = 'localhost'" "port = 15432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'postgres'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'scratch'" "" "[[postgres]]" "host = 'localhost'" "port = 15432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'postgres'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'old-replica'" > "$XDG_CONFIG_HOME/conm/default/postgres.toml" && clear`
Enter
Sleep 500ms
Show

Type "conm postgres"
Sleep 300ms
Enter
Sleep 1500ms
Type "j"
Sleep 600ms
Type "d"
Sleep 1500ms
Type "y"
Sleep 2000ms
Hide
Type "q"
Sleep 500ms
Type "clear"
Enter
Sleep 300ms
Show

Type "conm postgres remove old-replica"
Sleep 300ms
Enter
Sleep 1500ms
Type "conm postgres list"
Sleep 300ms
Enter
Sleep 2500ms
