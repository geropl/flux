package experimental_test


import "testing"
import "experimental"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,Sgf,DlXwgrw,2018-12-18T22:11:05Z,70
,,0,Sgf,DlXwgrw,2018-12-18T22:11:15Z,48
,,0,Sgf,DlXwgrw,2018-12-18T22:11:25Z,33
,,0,Sgf,DlXwgrw,2018-12-18T22:11:35Z,24
,,0,Sgf,DlXwgrw,2018-12-18T22:11:45Z,38
,,0,Sgf,DlXwgrw,2018-12-18T22:11:55Z,75

#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,1,Sgf,GxUPYq1,2018-12-18T22:11:05Z,96
,,1,Sgf,GxUPYq1,2018-12-18T22:11:15Z,-44
,,1,Sgf,GxUPYq1,2018-12-18T22:11:25Z,-25
,,1,Sgf,GxUPYq1,2018-12-18T22:11:35Z,46
,,1,Sgf,GxUPYq1,2018-12-18T22:11:45Z,-2
,,1,Sgf,GxUPYq1,2018-12-18T22:11:55Z,-14

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,2,Sgf,qaOnnQc,2018-12-18T22:11:05Z,-61.68790887989735
,,2,Sgf,qaOnnQc,2018-12-18T22:11:15Z,-6.3173755351186465
,,2,Sgf,qaOnnQc,2018-12-18T22:11:25Z,-26.049728557657513
,,2,Sgf,qaOnnQc,2018-12-18T22:11:35Z,114.285955884979
,,2,Sgf,qaOnnQc,2018-12-18T22:11:45Z,16.140262630578995
,,2,Sgf,qaOnnQc,2018-12-18T22:11:55Z,29.50336437998469
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_measurement,_field,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,0.3057387969724791
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,GxUPYq1,0.7564084754909545
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,qaOnnQc,0.6793279546139146
"
t_skew = (table=<-) =>
    table
        |> range(start: 2018-01-01T00:00:00Z)
        |> experimental.skew()

test _skew = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_skew})
