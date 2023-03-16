run: shrink
	go run main.go

download:
	yt-dlp -f 18 https://www.youtube.com/watch?v=FtutLA63Cp8

extract: download
	mkdir images
	ffmpeg -i ../【東方】Bad\ Apple!!\ ＰＶ【影絵】\ \[FtutLA63Cp8\].mp4 images/output_frames_%04d.png

shrink: extract
	mkdir images_small
	convert images/output_frames_*.png -resize 100x images/output_frames_%04d.png
