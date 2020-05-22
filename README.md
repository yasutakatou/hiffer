# hiffer

hiffer is **HISTORY + DIFF** tool, will be amazing and revolutional test automation tool!

## demo

(WIP)

## solution

When you want change human test operation to automation,  for example, you may use Ansible.<br>
Surely, can chane test operation to code. but, your team mate same?<br>
You will hit the wall.<br>
Your team mate must lern skill , know how , will spented  same learning cost, too.<br>

or when you must imediately operation occured (ex troubble), you direct operation on terminal. you must remember your operation, and will must remind operations when test .<br>

**hiffer!** this tool clear your worries.<br>
Can remember your operation  to history, and you check differerce anytime.<br>

## features

 - when operation on hiffer, write output to temporary file.
 - check diff by check sum, if diff exists, output colored text.
 - history can edittable, insert and delete.
 - convert strings pre execute, after execute. (for example, in case of command require tty)
 - file name auto completer in current directory 

## install

```
git clone https://github.com/yasutakatou/hiffer
cd hiffer
go build .
```

or download binary from release page.<br>
save binary file, copy to entryed execute path directory.<br>

## uninstall

```
delete that binary. del or rm command. (it's simple!)
```

## usecase

this tool wrap cmd.exe (when Windows),or unix shell. but you don't worry, you can usualy do operation.<br>
when you complete setting server and want to test setting,<br>
hiffer help you, do your operation again and diff output. so, you notice different value on setting.<br>

### this tool run option

options 
  -config string
        [-config=config file (default: .hiffer)] (default ".hiffer")
  -debug
        [-debug=debug mode (true is enable)]

-config this option set config file, and read it.
-debug is debug(verbose) mode. 

functions

when default input, this string execute on os and added history.
ex) ls
	execute "ls", and add output to history's end. 
note) this tool imprement tab completer, so you press tab, do file compeleter in current directory.

if another input , and api name the same, do following.

shell.AddCmd(&ishell.Cmd{Name: "@del",

delete history. you set history number to parameter, delete .
or use range parameter "-".
ex) @del 1-3
delete history number 1,2,3

shell.AddCmd(&ishell.Cmd{Name: "@env",

show enviroment value, or set enviroment value.

	TEST

TEST parameter is history number list for default test.
if you do @test no parameter,  use this value.
value is csv format and or use range parameter "-".
ex) @env TEST="1,3,5-10"
	test  to 1,3,5,6,7,8,9,10

note: if set "all", test all.

	SHELL

set use your favor shell . can use on linux only.
note: on windows, always using cmd.exe

	MAXHISTORYS

history max count. if you historys over this value, first history delete.
or historys over this value now,  delete historys until this value.

	PROMPT

this value is used prompt strings. 
note) when strings inclosed special character "`",  strings is execute and  setted to prompt.

ex) `pwd`
	your prompt => do `pwd`>

	TEMP

this value is temporary directory. use to write execute output.

shell.AddCmd(&ishell.Cmd{Name: "@export",

export config file. your parameter is file name.

shell.AddCmd(&ishell.Cmd{Name: "@ins",

execute command and insert output.
format is @ins (int) (command).
ex) @ins 5 "free"
execute "free" command and insert output to history number 5 .

shell.AddCmd(&ishell.Cmd{Name: "@kill",

all histories and output file delete in temporary directory.

note) if you ware changed temporary directory, old temprary directory not delete.

shell.AddCmd(&ishell.Cmd{Name: "@show",

show history list, or show old executed output.

ex) @show
show all list historys

ex) @show 1
show  history number 1 executed output

shell.AddCmd(&ishell.Cmd{Name: "@test",

execute history number, and to diff old output.
value is csv format and or use range parameter "-".
ex) @env TEST="1,3,5-10"
	test  to 1,3,5,6,7,8,9,10

note: if set "all", test all. and if parameter is empty, use enviroment value TEST.

shell.AddCmd(&ishell.Cmd{Name: "exit",

exit and save config.

	config file (defaut .hiffer)

basic, config detail can see @env section.
following option, CONVERT is another.

	CONVERT
	
this option is convert strings when pre execute command or after. 

using 4 values.
1 is convert keyword string. if strings include this value, converte strings. 
2 is when converting. "true" is pre execute, "false" is afier executed.
3 is  pre convert strings, 4 is after.

ex) 	vim,true,$,>&2
pre) vim test.txt
after) vim test.txt >&2

note) can use regex.
note) this config can edit on config file only. 
