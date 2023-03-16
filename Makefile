run: extract
	go run .

download:
	yt-dlp -f 18 https://www.youtube.com/watch?v=FtutLA63Cp8

extract: download
	if [ ! -d "images" ]; then \
		mkdir images; \
	fi

	ffmpeg -i 【東方】Bad\ Apple!!\ ＰＶ【影絵】\ \[FtutLA63Cp8\].mp4 images/output_frames_%04d.png

shrink: extract
	if [ ! -d "images_small" ]; then \
		mkdir images_small; \
	fi

	for i in images/*.png; do \
		convert $$i -resize 100x images_small/`basename $$i`; \
	done

shrink-big:
	if [ ! -d "images_small_big" ]; then \
		mkdir images_small; \
	fi

	for i in images/*.png; do \
		convert $$i -resize 200x images_small_big/`basename $$i`; \
	done
