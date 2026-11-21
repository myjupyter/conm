Output transport-use.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `D="$XDG_CONFIG_HOME/conm/default" && mkdir -p "$D" && printf "%s\n" "[[postgres]]" "host = 'postgres'" "port = 5432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'orders'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'orders'" > "$D/postgres.toml" && printf "%s\n" "[[ssh]]" "host = 'localhost'" "port = 2222" "username = 'demo'" "auth = 'password'" "password = 'demo-secret'" "" "[ssh.meta]" "name = 'bastion'" > "$D/ssh.toml" && printf "%s\n" "[[transport]]" "name = 'via-bastion'" "kind = 'ssh'" "via = 'bastion'" > "$D/transport.toml" && clear`
Enter
Sleep 500ms
Show

Type "conm postgres edit orders"
Sleep 300ms
Enter
Sleep 1500ms
Tab
Sleep 600ms
Type "l"
Sleep 600ms
Type "j"
Sleep 1000ms
Enter
Sleep 1500ms
Type "p"
Sleep 3000ms
Enter
Sleep 2500ms
Type "select inet_server_addr(), current_database();"
Sleep 300ms
Enter
Sleep 1800ms
Type "\q"
Sleep 300ms
Enter
Sleep 1500ms
Type "q"
Sleep 800ms
Type "conm postgres ping orders"
Sleep 300ms
Enter
Sleep 3500ms
