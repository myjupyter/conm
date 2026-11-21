Output settings-keybind.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `D="$XDG_CONFIG_HOME/conm/default" && mkdir -p "$D" && printf "%s\n" "[[postgres]]" "host = 'localhost'" "port = 15432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'orders'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'orders-local'" "description = 'Local orders database'" "tags = ['local', 'demo']" "" "[[postgres.meta.links]]" "name = 'grafana'" "url = 'https://grafana.example.com'" > "$D/postgres.toml" && clear`
Enter
Sleep 500ms
Show

Type "conm"
Sleep 300ms
Enter
Sleep 1500ms
Type "ss"
Sleep 1200ms
Tab
Sleep 1200ms
Type "jjjjjjjjjjjjjjjjjjj"
Sleep 1200ms
Type "a"
Sleep 1000ms
Type "o"
Sleep 600ms
Enter
Sleep 1500ms
Type "q"
Sleep 1200ms
Type "o"
Sleep 2500ms
Type "\q"
Sleep 300ms
Enter
Sleep 1200ms
Type "q"
Sleep 800ms
