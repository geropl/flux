package strings_test


import "testing"
import "strings"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:36Z,k9n  gm,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:46Z,b  .,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:06Z,cLnSkNMI  ī,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:16Z,13F2  z,used_percent,disk,disk1,apfs,host.local,/
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true,false
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path,sub
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,字,used_percent,disk,disk1,apfs,host.local,/,字
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k9n  gm,used_percent,disk,disk1,apfs,host.local,/,m
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b  .,used_percent,disk,disk1,apfs,host.local,/,.
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/,e
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,cLnSkNMI  ī,used_percent,disk,disk1,apfs,host.local,/,ī
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13F2  z,used_percent,disk,disk1,apfs,host.local,/,z
"
t_string_sub = (table=<-) =>
    table
        |> range(start: 2018-05-22T19:53:26Z)
        |> map(
            fn: (r) =>
                ({r with sub:
                        strings.substring(
                            v: r._value,
                            start: strings.strlen(v: r._value) - 1,
                            end: strings.strlen(v: r._value),
                        ),
                }),
        )

test _string_sub = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_string_sub})
