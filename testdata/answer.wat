(module
  (type (;0;) (func (result i32)))
  (import "test" "answer" (func $__imported_answer (type 0)))
  (func $answer (type 0) (result i32)
    call $__imported_answer)
  (export "answer" (func $answer)))
