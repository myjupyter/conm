Output exec-shell.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `cd "$(mktemp -d)" && mkdir -p "$XDG_CONFIG_HOME/conm/default" && printf "%s\n" "[[postgres]]" "host = 'localhost'" "port = 15432" "username = 'demo'" "password = 'demo-secret'" "dbname = 'orders'" "sslmode = 'disable'" "" "[postgres.meta]" "name = 'orders-local'" > "$XDG_CONFIG_HOME/conm/default/postgres.toml"`
Enter
Type `printf "%s\n" "create table if not exists items (id int, title text);" "insert into items values (1, 'book'), (2, 'lamp');" "select count(*) as items from items;" > script.sql && clear`
Enter
Sleep 500ms
Show

Type 'conm postgres exec orders-local -- "select current_user, current_database();"'
Sleep 300ms
Enter
Sleep 1800ms
Type "cat script.sql"
Sleep 300ms
Enter
Sleep 1500ms
Type "conm postgres orders-local < script.sql"
Sleep 300ms
Enter
Sleep 2000ms
Type 'conm postgres exec orders-local -- "select nope;"; echo "exit code $?"'
Sleep 300ms
Enter
Sleep 2500ms
