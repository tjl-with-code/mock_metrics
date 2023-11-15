package main

import (
	"bytes"
	"fake-metrics/metric_data"
	"fmt"
	"log"
	"testing"
)

func TestA(t *testing.T) {
	ss := `
mysql_global_status_innodb_row_lock_waits{app="mysql-exporter",port="3306",ip="58.222.126.1"} 54.0
mysql_global_variables_innodb_ft_cache_size{app="mysql-exporter",port="3306",ip="58.222.126.1"} 8000000.0
mysql_global_variables_innodb_purge_threads{app="mysql-exporter",port="3306",ip="58.222.126.1"} 4.0
mysql_global_status_key_blocks_unused{app="mysql-exporter",port="3306",ip="58.222.126.1"} 6696.0
mysql_global_variables_innodb_change_buffer_max_size{app="mysql-exporter",port="3306",ip="58.222.126.1"} 25.0
mysql_global_variables_innodb_ft_total_cache_size{app="mysql-exporter",port="3306",ip="58.222.126.1"} 6.4E8
mysql_global_status_uptime{app="mysql-exporter",port="3306",ip="58.222.126.1"} 55768.0
mysql_exporter_scrape_errors_total{app="mysql-exporter",port="3306",ip="58.222.126.1",collector="collect.info_schema.innodb_cmp"} 239831.0
mysql_exporter_scrape_errors_total{app="mysql-exporter",port="3306",ip="58.222.126.1",collector="collect.info_schema.innodb_cmpmem"} 239831.0
mysql_exporter_scrape_errors_total{app="mysql-exporter",port="3306",ip="58.222.126.1",collector="collect.slave_status"} 239831.0
`
	bf := bytes.NewBufferString(ss)
	bf.WriteByte('\n')
	mm, err := metric_data.NewTextParser().Encode(bf)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(mm)
}
