FROM busybox:musl
ADD cleanup/cleanup /
ADD filescan/filescan /
ADD thumbnailer/thumbnailer /
CMD [ "/bin/sh", "-c", "echo Running Cleanup; /cleanup; echo Running Filescan; /filescan; echo Running thumbnailer; /thumbnailer"]
