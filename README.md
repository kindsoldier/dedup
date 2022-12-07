# DeDup


```

NAME
       dedup - search for files with the same content

SYNOPSIS
       dedup [option] dirs...


DESCRIPTION
       Utility for finding files with the same content and creating hard links
       to the first copy

       The utility is written to find duplicates in the archive of photos  and
       videos.   If  when  reading the next file the hash sum and size are the
       same as the previous one, it is considered  a  duplicate.   File  time‐
       stamps are not taken into account.

       To  reduce  the  search time in a large archive you can specify to read
       only the first N bytes, usually different video  files  have  different
       checksums  already within the first 1Mb.  If the file sizes are differ‐
       ent, the file will also not be considered a duplicate.

       By default, only the listing is created.

       Existing hard link is not counted because it may lead outside the  tree
       it is analyzing.


OPTIONS
       Usage: dedup [option] dirs...  -bak rename dumplicate before link

       -depth int
              maximal depth (default 5)

       -limit int
              read only first bytes (default "0b")

       -max string
              maximal size (default "1Gb")

       -min string
              minimal size (default "1kb")

       -pat string
              file name pattern (default "*")

       -quiet supress listing

       -stats print addtional stats

EXIT STATUS
       edup  exits  with  status  0  if  all files are processed successfully,
       greater than 0 if errors occur.

```
