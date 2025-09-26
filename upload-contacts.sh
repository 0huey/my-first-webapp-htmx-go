curl -s -c cookie.txt -o /dev/null http://127.0.0.1:8080/login -d "username=admin&password=admin"
grep -v "'" /usr/share/dict/words | sort -R | head -n 1000 | while read name; do
	curl -s -o /dev/null -b cookie.txt 127.0.0.1:8080/contacts -d "name=$name&email=$name@example.com"
done
rm cookie.txt
