# Audio Library Tools

## Checker

Check some rules to detect audio files tags inconsistencies:

- missing album name, track number or title
- track with diffrent album into same directory
- same track number into same directory
- title, artist, or album contains suspicious word (untitled, track, unknow)


```bash
./audio-lib-tools check --tracks --albums --only-errrors ~/Music
```

## Exporter

Export directory audio files to json


```bash
./audio-lib-tools export  ~/Music/ ~/Music/export.json
```

## License

See ```LICENSE``` for more information