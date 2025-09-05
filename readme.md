Learning to write a web app with HTMX and Go, some what following along with ThePrimagens Front End Masters course

https://theprimeagen.github.io/fem-htmx

https://www.youtube.com/watch?v=x7v6SNIgJpE

add some sample contact data
```
grep -v "'" /usr/share/dict/words | sort -R | head -n 1000 | while read name; do curl -s localhost:8080/contacts -d "name=$name&email=$name@example.com" >/dev/null; done
```
