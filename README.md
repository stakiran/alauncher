# alauncher
Create your alias with the single ini file for Windows.

## Requirement
- Golang
- "github.com/stretchr/testify/assert" for unittest
- "github.com/go-ini/ini" for ini operation

## Installation

```
$ go get github.com/stakiran/alauncher
```

If you want to use an alias of alauncher, like this:

```
[ala]
rawbin=alauncher %*
```

## Usage

### Firstly, create the ini file.

```
$ alauncher
The ini file 'C:\Users\XXXXXXXX\.ala.ini' not found, so create firstly.
```

### Secondly, edit the ini file.
- Add a directory path passed to PATH to `outdir`.
- Add your aliaes.

### Build and so on.
After editing, Do `alauncher` command to create aliases from the ini file.

```
$ alauncher
```

You can also use dryrun.

```
$ alauncher -dryrun
```

Edit the ini file with your editor.

```
$ alauncher -edit
```

## About the ini file
See also your first .ala.ini or sample [.ala.ini](.ala.ini).

### Nmae and location
The ini file name is `.ala.ini`, and this is created on %HOME% or %USERPROFILE% directory.

### Section
There are three type of section: options, variables and alias.

### Section > options

```
[_options]
outdir=D:\bin1\alauncher
```

You must set a directory path passed to PATH to `outdir`.

### Section > variables

```
[_variables]
sys32=%windir%\system32
aladir=D:\bin1\alauncher
hidemaru=C:\Program Files (x86)\Hidemaru\HIDEMARU.EXE
conemu=C:\Program Files\ConEmu\ConEmu64.exe
```

If need, you can define variables with `key=value` format.

### Section > alias > rawbin
[Example] `pd` command as the alias of `pushd`.

Ini:

```
[pd]
rawbin=pushd %*
disable=setlocal,echooff

[pd1]
rawbin=pushd %*
disable=setlocal,echooff
```

Generated batches:

```
$ type pd.bat




pushd %*
$ type pd1.bat
@echo off

setlocal

pushd %*
```

### Section > alias > bin
[Example] `hide` and `hidemaru` command as the alias of my favorite text editor "Hidemaru Editor".

Ini:

```
[_variables]
hidemaru=C:\Program Files (x86)\Hidemaru\HIDEMARU.EXE

[hidemaru]
bin=%hidemaru%
prm=%*
alias=hide
```

Generated batched:

```
$ type hidemaru.bat
@echo off

setlocal

pushd %cd%

start "" "C:\Program Files (x86)\Hidemaru\HIDEMARU.EXE" %*

popd
$ type hide.bat
@echo off

call %~dp0hidemaru.bat %*

```

### Section > alias > separator
If you have many aliases, you can use a separator alias as a readable section.

```
[____CUIAliaes____]
ignore_this=true
```

Use `ignore_this=true`, do not create the batch file.

## FAQ

## Q: Is it possible that use prefix or suffix spaces?
Ans: possible.

Use `%s%` variable. This is alauncher's system variable.

Ini:

```
[pp]
rawbin=prompt $$%s%
disable=setlocal
```

pp.bat:

```
$ type pp.bat
@echo off



prompt $$ 
```

Use:

```
C:\Users\XXXXXXX>pp

$ echo Yeah!
Yeah!

$ 
```

## How to develop
Call run.bat to run.

Call test.bat to test.

Call build.bat to build.

test_alias.ini is one of the test data.

## License
[MIT License](LICENSE)

## Author
[stakiran](https://github.com/stakiran)
