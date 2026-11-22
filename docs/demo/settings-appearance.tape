Output settings-appearance.gif

Set Shell "bash"
Set FontSize 16
Set Width 1280
Set Height 760
Set Padding 12
Set Framerate 20
Set TypingSpeed 45ms
Set Theme "Catppuccin Mocha"
Env COLORTERM "truecolor"

Hide
Type `D="$XDG_CONFIG_HOME/conm/default" && mkdir -p "$D" && printf "%s\n" "[[postgres]]" "host = 'localhost'" "port = 15432" "username = 'demo'" "dbname = 'orders'" "[postgres.meta]" "name = 'orders-local'" "[[postgres]]" "host = 'db.billing.internal'" "port = 5432" "username = 'billing_ro'" "dbname = 'billing'" "[postgres.meta]" "name = 'billing-prod'" "[[postgres]]" "host = 'auth.staging.internal'" "port = 5432" "username = 'auth'" "dbname = 'auth'" "[postgres.meta]" "name = 'auth-staging'" > "$D/postgres.toml" && clear`
Enter
Sleep 500ms
Show

Type "conm"
Sleep 300ms
Enter
Wait+Screen /connections/
Sleep 1500ms
Type "ss"
Wait+Screen /appearance/
Sleep 1000ms
Tab
Sleep 400ms
Tab
Sleep 800ms
Enter
Sleep 1000ms
Right
Sleep 700ms
Right
Sleep 700ms
Right
Sleep 900ms
Type "i"
Sleep 400ms
Type "blue>pink"
Sleep 500ms
Enter
Sleep 1500ms
Tab
Sleep 400ms
Tab
Sleep 900ms
Right
Sleep 700ms
Right
Sleep 1200ms
Tab
Sleep 900ms
Right
Sleep 1000ms
Down
Sleep 500ms
Right
Sleep 250ms
Right
Sleep 250ms
Right
Sleep 250ms
Right
Sleep 250ms
Right
Sleep 250ms
Right
Sleep 1500ms
Type "q"
Wait+Screen /connections/
Sleep 2500ms
Type "q"
Sleep 500ms
