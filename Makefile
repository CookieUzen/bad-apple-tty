build:
	go build -o .

download:
	yt-dlp -f 18 https://www.youtube.com/watch?v=FtutLA63Cp8

run: bad-apple-tty download
	./bad-apple-tty '【東方】Bad Apple!! ＰＶ【影絵】 [FtutLA63Cp8].mp4'
