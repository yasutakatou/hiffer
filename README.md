# hiffer

hiffer is **HISTORY + DIFF** tool, will be amazing and revolutional test automation tool!

## demo

![demo1](https://github.com/yasutakatou/hiffer/blob/pic/hiffer.gif)

support linux and windows.<br>

![demo2](https://github.com/yasutakatou/hiffer/blob/pic/windows.png)

## solution

When you want change human test operation to automation,  for example, you may use Ansible.<br>
Surely, can chane test operation to code. but, your team mate same?<br>
<br>
*You will hit the wall.*<br>
Your team mate must lern skill , know how , will spented  same learning cost, too.<br>
<br>
or when you must imediately operation occured (ex troubble), you direct operation on terminal.<br>
You must remember your operation, and will must remind operations when test .<br>
<br>
**hiffer!** this tool clear your worries.<br>
Can remember your operation  to history, and you check differerce anytime.<br>

## features

 - When operation on hiffer, write output to temporary file.
 - Check diff by check sum, if diff exists, *output colored text*.
 - History can *edittable*, insert and delete.
 - Convert strings pre execute, after execute. (for example, in case of command require tty)
 - Suppoert *file name auto completer* in current directory 

## install

```
git clone https://github.com/yasutakatou/hiffer
cd hiffer
go build .
```

or download binary from [release page](https://github.com/yasutakatou/hiffer/releases).<br>
save binary file, copy to entryed execute path directory.<br>

## uninstall

```
delete that binary.
```

del or rm command. *(it's simple!)*

## usecase

This tool *wrap cmd.exe (when Windows)*,or *unix shell*. but you don't worry, <br>
You can usualy do operation. when you complete server and want to test setting,<br>
**hiffer help you, do your operation again and diff output. so, you notice different value on setting.**<br>

### this tool run option

options<br>

```
-config strings [-config=config file (default: .hiffer)] (default ".hiffer")
    -config this option set config file, and read it.
-debug [-debug=debug mode (true is enable)]
    -debug is debug(verbose) mode. 
```

### functions

When default input, this string execute on os and added history.<br>

ex) ls<br>
execute "ls", and add output to history's end. <br>

note) This tool **imprement tab completer**, so you press tab, do file compeleter in current directory.<br>

![demo1](https://github.com/yasutakatou/hiffer/blob/pic/completer.gif)

note) if add to head character "!", not add historys.<br>
<br>
If another input , and api name the same, do following.<br>
<br>

#### @del

delete history. you set history number to parameter, delete. or use range parameter "-".<br>
<br>
ex) @del 1-3<br>
delete history number 1,2,3<br>

#### @env

show enviroment value, or set enviroment value.<br>
You can set following environment parameter.<br>

	TEST

TEST parameter is history number list for default test. if you do @test no parameter, use this value.<br>
value is csv format and or use range parameter "-".<br>

ex) @env TEST="1,3,5-10"<br>
test  to 1,3,5,6,7,8,9,10<br>
<br>
note: if set "all", test all.<br>

	SHELL

set use your favor shell . can use on linux only.<br>
note: on windows, always using cmd.exe<br>

	MAXHISTORYS

history max count. if you historys over this value, first history delete.<br>
or historys over this value now,  delete historys until this value.<br>

	PROMPT

This value is used prompt strings. <br>
note) when strings inclosed special character "\`",  strings is execute and  setted to prompt.<br>
note) strings execute on os by your privilege. so, don't use dangerous command.<br>

ex) \`pwd\`<br>
your prompt => do \`pwd\`><br>

	TEMP

This value is temporary directory. use to write execute output.<br>

#### @export

export config file. your parameter is file name.<br>

#### @ins

execute command and insert output. format is @ins (int) (command).<br>
<br>
ex) @ins 5 "free"<br>
execute "free" command and insert output to history number 5.<br>

#### @kill

all histories and output file delete in temporary directory.<br>
<br>
note) if you ware changed temporary directory, old temprary directory not delete.<br>

#### @show

show history list, or show old executed output.<br>
<br>
ex) @show<br>
show all list historys<br>
<br>
ex) @show 1<br>
show  history number 1 executed output<br>

#### @test

execute history number, and to diff old output. value is csv format and or use range parameter "-".<br>
<br>
ex) @env TEST="1,3,5-10"<br>
test  to 1,3,5,6,7,8,9,10<br>
<br>
note: if set "all", test all. and if parameter is empty, use enviroment value TEST.<br>

#### exit

exit and save config.<br>

## config file (defaut .hiffer)

basic, config detail can see @env section. following option, CONVERT is another.<br>

	CONVERT
	
This option is convert strings when pre execute command or after. <br>
<br>
using **4** values.<br>
**1** is convert keyword string. if strings include this value, converte strings. <br>
**2** is when converting. *"true"* is pre execute, *"false"* is afier executed.<br>
**3** is  pre convert strings, **4** is after.<br>
<br>
ex) vim,true,$,>&2<br>
pre) vim test.txt<br>
after) vim test.txt **>&2**<br>
<br>
note) **can use regex.**<br>
note) this config can edit on config file only. <br>
