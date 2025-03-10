// Package schema provides functions for exploring your InfluxDB data schema.
//
// introduced: 0.88.0
package schema


// fieldsAsCols is a special application of `pivot()` that pivots input data
// on `_field` and `_time` columns to align fields within each input table that
// have the same timestamp.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Pivot InfluxDB fields into columns
// ```
// # import "array"
// import "influxdata/influxdb/schema"
//
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "temp", _value: "73.1"},
// #         {_time: 2021-01-02T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "temp", _value: "68.2"},
// #         {_time: 2021-01-03T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "temp", _value: "61.4"},
// #         {_time: 2021-01-01T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "hum", _value: "89.2"},
// #         {_time: 2021-01-02T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "hum", _value: "90.5"},
// #         {_time: 2021-01-03T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "hum", _value: "81.0"},
// #     ],
// # )
// #     |> group(columns: ["_time", "_value"], mode: "except")
// #
// < data
// >     |> schema.fieldsAsCols()
// ```
//
// tags: transformations
//
fieldsAsCols = (tables=<-) =>
    tables
        |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")

// tagValues returns a list of unique values for a given tag.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to return unique tag values from.
// - tag: Tag to return unique values from.
// - predicate: Predicate function that filters tag values.
//   Default is `(r) => true`.
// - start: Oldest time to include in results. Default is `-30d`.
//
//   Relative start times are defined using negative durations.
//   Negative durations are relative to `now()`.
//   Absolute start times are defined using time values.
//
// ## Examples
//
// ### Query unique tag values from an InfluxDB bucket
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.tagValues(
//     bucket: "example-bucket",
//     tag: "host",
// )
// ```
//
// tags: metadata
//
tagValues = (bucket, tag, predicate=(r) => true, start=-30d) =>
    from(bucket: bucket)
        |> range(start: start)
        |> filter(fn: predicate)
        |> keep(columns: [tag])
        |> group()
        |> distinct(column: tag)

// tagKeys returns a list of tag keys for all series that match the `predicate`.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to return tag keys from.
// - predicate: Predicate function that filters tag keys.
//   Default is `(r) => true`.
// - start: Oldest time to include in results. Default is `-30d`.
//
//   Relative start times are defined using negative durations.
//   Negative durations are relative to `now()`.
//   Absolute start times are defined using time values.
//
// ## Examples
//
// ### Query tag keys in an InfluxDB bucket
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.tagKeys(bucket: "example-bucket")
// ```
//
// tags: metadata
//
tagKeys = (bucket, predicate=(r) => true, start=-30d) =>
    from(bucket: bucket)
        |> range(start: start)
        |> filter(fn: predicate)
        |> keys()
        |> keep(columns: ["_value"])
        |> distinct()

// measurementTagValues returns a list of tag values for a specific measurement.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to return tag values from for a specific measurement.
// - measurement: Measurement to return tag values from.
// - tag: Tag to return all unique values from.
//
// ## Examples
//
// ### Query unique tag values from an InfluxDB measurement
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.measurementTagValues(
//     bucket: "example-bucket",
//     measurement: "example-measurement",
//     tag: "example-tag",
// )
// ```
//
// tags: metadata
//
measurementTagValues = (bucket, measurement, tag) =>
    tagValues(bucket: bucket, tag: tag, predicate: (r) => r._measurement == measurement)

// measurementTagKeys returns the list of tag keys for a specific measurement.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to return tag keys from for a specific measurement.
// - measurement: Measurement to return tag keys from.
//
// ## Examples
//
// ### Query tag keys from an InfluxDB measurement
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.measurementTagKeys(
//     bucket: "example-bucket",
//     measurement: "example-measurement",
// )
// ```
//
// tags: metadata
//
measurementTagKeys = (bucket, measurement) => tagKeys(bucket: bucket, predicate: (r) => r._measurement == measurement)

// fieldKeys returns field keys in a bucket.
//
// Results include a single table with a single column, `_value`.
//
// **Note**: FieldKeys is a special application of `tagValues that returns field
// keys in a given bucket.
//
// ## Parameters
// - bucket: Bucket to list field keys from.
// - predicate: Predicate function that filters field keys.
//   Default is `(r) => true`.
// - start: Oldest time to include in results. Default is `-30d`.
//
//   Relative start times are defined using negative durations.
//   Negative durations are relative to `now()`.
//   Absolute start times are defined using time values.
//
// ## Examples
//
// ### Query field keys from an InfluxDB bucket
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.fieldKeys(bucket: "example-bucket")
// ```
//
// tags: metadata
//
fieldKeys = (bucket, predicate=(r) => true, start=-30d) =>
    tagValues(bucket: bucket, tag: "_field", predicate: predicate, start: start)

// measurementFieldKeys returns a list of fields in a measurement.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to retrieve field keys from.
// - measurement: Measurement to list field keys from.
// - start: Oldest time to include in results. Default is `-30d`.
//
//   Relative start times are defined using negative durations.
//   Negative durations are relative to `now()`.
//   Absolute start times are defined using time values.
//
// ## Examples
//
// ### Query field keys from an InfluxDB measurement
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.measurementFieldKeys(
//     bucket: "example-bucket",
//     measurement: "example-measurement",
// )
// ```
//
// tags: metadata
//
measurementFieldKeys = (bucket, measurement, start=-30d) =>
    fieldKeys(bucket: bucket, predicate: (r) => r._measurement == measurement, start: start)

// measurements returns a list of measurements in a specific bucket.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to retrieve measurements from.
//
// ## Examples
//
// ### Return a list of measurements in an InfluxDB bucket
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.measurements(bucket: "example-bucket")
// ```
//
// tags: metadata
//
measurements = (bucket) => tagValues(bucket: bucket, tag: "_measurement")
