Output import-interactive.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"

Hide
Type `mkdir -p "$XDG_CONFIG_HOME" && export PGSERVICEFILE="$XDG_CONFIG_HOME/pg_service.conf" PGPASSFILE="$XDG_CONFIG_HOME/pgpass" && printf "%s\n" "[orders-local]" "host=localhost" "port=15432" "user=demo" "dbname=orders" "" "[orders-staging]" "host=staging.internal" "user=demo" "dbname=orders" "sslmode=require" "" "[billing]" "host=billing.internal" > "$PGSERVICEFILE" && clear`
Enter
Sleep 800ms
Show

Type "conm postgres import --show"
Sleep 300ms
Enter
Sleep 3000ms
Type "conm postgres import -i"
Sleep 300ms
Enter
Sleep 3000ms

Type "jj"
Sleep 400ms
Type "i"
Sleep 120ms
Type "demo"
Sleep 150ms
Enter
Sleep 300ms
Type "jj"
Sleep 400ms
Type "i"
Sleep 120ms
Type "billing"
Sleep 150ms
Enter
Sleep 800ms

Tab
Sleep 600ms
Tab
Sleep 1000ms
Tab
Sleep 500ms
Enter
Sleep 3000ms
Type "conm postgres list"
Sleep 300ms
Enter
Sleep 2500ms
