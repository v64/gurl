Gurl: A quick, multithreaded URL downloader.
License: Public domain

To run: go run gurl.go

Expects a CSV in the same directory named urls.csv that looks like this:
"http://www.example.com/img/img1.jpg"
"http://www.example2.com/img2.jpg"
"http://host1.www.examples.com/images/img/imag2.jpeg"

and downloads the files and saves them this way:
./output/www.example.com/httpwwwexamplecomimgimg1.jpg
./output/www.example2.com/httpwwwexample2comimg2.jpg
./output/host1.www.examples.com/httphost1wwwexamplescomimagesimgimag2.jpeg