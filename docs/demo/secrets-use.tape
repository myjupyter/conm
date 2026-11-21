Output secrets-use.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `D="$XDG_CONFIG_HOME/conm/default" && mkdir -p "$D" "$XDG_DATA_HOME/conm/default" && printf "%s\n" "[[postgres]]" "host = 'localhost'" "port = 15432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'orders'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'orders-local'" > "$D/postgres.toml" && printf "%s\n" "workspace_id = '00000000-0000-4000-8000-00000000c0de'" "" "[[keyring]]" "id = 'orders-rw'" "provider = 'keyring'" "location = 'orders'" "description = 'Orders read-write user'" > "$XDG_DATA_HOME/conm/default/secret.toml" && security add-generic-password -U -s conm.00000000-0000-4000-8000-00000000c0de -a orders -w demo-secret && clear`
Enter
Sleep 800ms
Show

Type "conm postgres edit orders-local"
Sleep 300ms
Enter
Sleep 1500ms
Type "jjj"
Sleep 600ms
Type "l"
Sleep 500ms
Type "l"
Sleep 1000ms
Type "s"
Sleep 2000ms
Enter
Sleep 1500ms
Enter
Sleep 1500ms
Type "p"
Sleep 2000ms
Type "q"
Sleep 800ms
Type 'grep password "$XDG_CONFIG_HOME/conm/default/postgres.toml"'
Sleep 300ms
Enter
Sleep 2500ms
