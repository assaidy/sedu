## sedu: search duplicate files

### Build

```shell
$ git clone https://github.com/assaidy/sedu
$ cd ./sedu
$ go build
```
### Usage

```shell
$ ./sedu <dir-path>
```

### Sample output testing on ~/Documents
```txt
[INFO] collecting all files...
[INFO] generating file hashes...
[INFO] printing all duplicate files...
{
   /home/me/Documents/xxx
   /home/me/Documents/y/xxx
}
{
   /home/me/Documents/programming/cpp/xxx
   /home/me/Documents/programming/dsa/sorting/xxx
}
{
   ...
   ...
   ...
   ...
}
```
