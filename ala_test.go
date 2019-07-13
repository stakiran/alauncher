package main

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/go-ini/ini"
)

func TestUtils(t *testing.T) {
	// \ / : * ? " < > |
	assert.False(t, isInvalidFilename("hoge"))
	assert.False(t, isInvalidFilename("あいうえお"))
	assert.True(t, isInvalidFilename("?"))
	assert.True(t, isInvalidFilename("*"))
	assert.True(t, isInvalidFilename("<section>"))
	assert.True(t, isInvalidFilename("sec:tion"))
	assert.False(t, isInvalidFilename("semicolon is ok;"))
}

func TestCommand(t *testing.T) {
	a := NewCommand()
	assert.Equal(t, a.Name, "")
	assert.Equal(t, a.Rawbin, "")
	assert.Equal(t, a.Bin, "")
	assert.Equal(t, a.Dir, "")
	assert.Equal(t, a.Prm, "")
	assert.Equal(t, len(a.Aliases), 0)
	assert.Equal(t, a.IsIgnored, false)
	assert.Equal(t, a.DisableSetlocal, false)

	cfg, err := ini.Load("test_alias.ini")
	if err != nil {
		abort("Fail to read in TestCommand.")
	}

	sec := cfg.Section("bat_for_test")
	a.ImportFromGoIni(sec)
	assert.Equal(t, a.Name, "bat_for_test")
	assert.Equal(t, a.Rawbin, "notepad")
	assert.Equal(t, a.Bin, `c:\windows\notepad.exe`)
	assert.Equal(t, a.Dir, "cur")
	assert.Equal(t, a.Prm, "notfoundfile.txt")
	assert.Equal(t, len(a.Aliases), 2)
	assert.Equal(t, a.Aliases[0], "bft")
	assert.Equal(t, a.Aliases[1], "batfortest")
	assert.Equal(t, a.IsIgnored, true)
	assert.Equal(t, a.DisableSetlocal, true)
}

func TestOption(t *testing.T) {
	cfg, err := ini.Load("test_alias.ini")
	if err != nil {
		abort("Fail to read in TestAlias.")
	}

	option := NewOption(cfg)
	assert.Equal(t, option.Outdir, ".")
}

func TestVariableDeployer(t *testing.T) {
	cfg, err := ini.Load("test_alias.ini")
	if err != nil {
		abort("Fail to read in TestAlias.")
	}

	deployer := NewVariableDeployer(cfg)

	assert.Equal(t, deployer.Deploy(`%programdir%\mozilla`), `c:\program files\mozilla`)
	assert.Equal(t, deployer.Deploy(`explorer %wd%`), `explorer c:\windows`)
	assert.Equal(t, deployer.Deploy(`notepad.exe %up_with_CamelCase%\.gitconfig`), `notepad.exe %UserProfile%\.gitconfig`)
	assert.Equal(t, deployer.Deploy(`%pydir%\python.exe %pydir%\Lib\trace.py`), `D:\bin\Python361\python.exe D:\bin\Python361\Lib\trace.py`)
	assert.Equal(t, deployer.Deploy(`%multiple1%`), `c:\windows`)
	assert.Equal(t, deployer.Deploy(`%multiple2%`), `c:\windows\system32`)
	assert.Equal(t, deployer.Deploy(`%multiple_reverse_order%`), `c:\windows\system32\drivers\etc\hosts`)

	// Do not occur circular reference
	deployer.Deploy(`%circular1%`)
	deployer.Deploy(`%circular2%`)
	deployer.Deploy(`%circular3%`)
}

func TestVariableDeployerAlaVaribles(t *testing.T) {
	cfg, err := ini.Load("test_alias.ini")
	if err != nil {
		abort("Fail to read in TestAlias.")
	}

	deployer := NewVariableDeployer(cfg)

	assert.NotEqual(t, deployer.Deploy(`prompt $$%s%`), `prompt $$ `)
	assert.Equal(t, deployer.Deploy(`prompt $$%s%`), `prompt $$%s%`)
	// auto-trimmed by go-ini.
	assert.Equal(t, deployer.Deploy(`%suffixspace%`), `hoge`)
	assert.Equal(t, deployer.Deploy(`%prefixspace%`), `hoge`)
	assert.Equal(t, deployer.Deploy(`%prefixsuffixspace%`), `hoge`)
	assert.Equal(t, deployer.Deploy(`%prefixsuffixspace_with_alavar_s%`), `%s%hoge%s%`)

	deployer.AddAlaVariables()

	assert.Equal(t, deployer.Deploy(`prompt $$%s%`), `prompt $$ `)
	assert.Equal(t, deployer.Deploy(`%prefixsuffixspace_with_alavar_s%`), ` hoge `)
}

func TestGenerator(t *testing.T) {
	cfg, err := ini.Load("test_alias.ini")
	if err != nil {
		abort("Fail to read in TestAlias.")
	}

	deployer := NewVariableDeployer(cfg)

	// case1:
	// rawbin with disabling setlocal

	sec := cfg.Section("bat_disable")
	command := NewCommand()
	command.ImportFromGoIni(sec)
	gen := NewGenerator(command, deployer)
	actual := gen.Generate()
	expect := `@echo off



set ALA_VERSION=x.y.z`
	assert.Equal(t, actual, expect)

	// case2:
	// bin dir prm  valid?
	// o   o   o    y

	sec = cfg.Section("bat_exec_givendir")
	command = NewCommand()
	command.ImportFromGoIni(sec)
	gen = NewGenerator(command, deployer)
	actual = gen.Generate()
	expect = `@echo off

setlocal

pushd %userprofile%

start "" "c:\windows\notepad.exe" .gitconfig

popd`
	assert.Equal(t, actual, expect)

	// case3:
	// bin dir prm  valid?
	// o   x   o    y       dir is completed with 'cur'

	sec = cfg.Section("bat_exec_bindir")
	command = NewCommand()
	command.ImportFromGoIni(sec)
	gen = NewGenerator(command, deployer)
	actual = gen.Generate()
	expect = `@echo off

setlocal

pushd c:\windows

start "" "c:\windows\notepad.exe" system.ini

popd`
	assert.Equal(t, actual, expect)

}
