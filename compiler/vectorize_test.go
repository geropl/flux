package compiler_test

// func TestVectorizedFunctions(t *testing.T) {
//   testCases := []struct {
//     name           string
//     fn             string
//     inType         semantic.MonoType
//     input          values.Object
//     want           values.Object
//     wantCompileErr bool
//     wantEvalErr    bool
//   }{
//     {
//       name: "map identity fn",
//       fn:   `(r) => ({r with a: r.a, b: r.b})`,
//       inType: semantic.NewObjectType([]semantic.PropertyType{
//         {Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
//           {Key: []byte("a"), Value: semantic.NewVectorType(semantic.BasicInt)},
//           {Key: []byte("b"), Value: semantic.NewVectorType(semantic.BasicInt)},
//         })},
//       }),
//       input: []values.Object{
//         values.NewObjectWithValues(map[string]values.Value{
//           "a": values.NewInt(1),
//           "b": values.NewInt(11),
//         }),
//         values.NewObjectWithValues(map[string]values.Value{
//           "a": values.NewInt(2),
//           "b": values.NewInt(12),
//         }),
//         values.NewObjectWithValues(map[string]values.Value{
//           "a": values.NewInt(3),
//           "b": values.NewInt(13),
//         }),
//         values.NewObjectWithValues(map[string]values.Value{
//           "a": values.NewInt(4),
//           "b": values.NewInt(14),
//         }),
//       },
//       want: values.NewObjectWithValues(map[string]values.Value{
//         "a": arrowutil.NewVectorFromSlice([]values.Value{
//           values.NewInt(int64(1)),
//           values.NewInt(int64(2)),
//           values.NewInt(int64(3)),
//           values.NewInt(int64(4)),
//         }, flux.TInt),
//         "b": arrowutil.NewVectorFromSlice([]values.Value{
//           values.NewInt(int64(11)),
//           values.NewInt(int64(12)),
//           values.NewInt(int64(13)),
//           values.NewInt(int64(14)),
//         }, flux.TInt),
//       }),
//     },
//   }
//   for _, tc := range testCases {
//     t.Run(tc.name, func(t *testing.T) {
//       pkg, err := runtime.AnalyzeSource(tc.fn)
//       if err != nil {
//         t.Fatalf("unexpected error: %s", err)
//       }
//
//       stmt := pkg.Files[0].Body[0].(*semantic.ExpressionStatement)
//       fn := stmt.Expression.(*semantic.FunctionExpression)
//       f, err := compiler.Compile(nil, fn, tc.inType)
//       if err != nil {
//         if !tc.wantCompileErr {
//           t.Fatalf("unexpected error: %s", err)
//         }
//         return
//       } else if tc.wantCompileErr {
//         t.Fatal("wanted error but got nothing")
//       }
//
//       got, err := f.Eval(context.TODO(), tc.input)
//       if tc.wantEvalErr != (err != nil) {
//         t.Errorf("unexpected error: %s", err)
//       }
//
//       if !cmp.Equal(tc.want, got, CmpOptions...) {
//         t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(tc.want, got, CmpOptions...))
//       }
//     })
//   }
// }
