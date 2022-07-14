from __future__ import annotations
from dataclasses import dataclass, field, replace
from textwrap import dedent
import io


class Printable:
    def __repr__(self):
        return str(self)

    def __str__(self):
        stream = io.StringIO()
        self.to_str(stream, {}, 0)
        return stream.getvalue()

    def to_str(self, stream: io.TextIOBase, seen: dict, level: int):
        seen[id(self)] = len(seen)
        stream.write(f"<#{seen[id(self)]} ")
        stream.write(type(self).__name__)

        def indent(s, level):
            stream.write(s + "\n" + "  "*level)

        def write_value(value, level):
            if id(value) in seen:
                stream.write(f"<#{seen[id(value)]}>")
                return
            if value is None:
                stream.write("nil")
            elif isinstance(value, Printable):
                seen[id(value)] = len(seen)
                value.to_str(stream, seen, level+1)
            elif isinstance(value, (int, str, float, bool)):
                stream.write(str(value))
            elif isinstance(value, (list, set, tuple)):
                seen[id(value)] = len(seen)
                indent("[", level+1)
                for i, x in enumerate(value):
                    if i:
                        indent(",", level+1)
                    write_value(x, level+1)
                indent("]", level)
            elif isinstance(value, dict):
                seen[id(value)] = len(seen)
                indent("{", level+1)
                for i, (k, v) in enumerate(value.items()):
                    if i:
                        indent(",", level+1)
                    write_value(k, indent+1)
                    stream.write(": ")
                    write_value(v, indent+1)
                stream.write("}", level)
            else:
                stream.write("...")

        count = 0
        for name in dir(self):
            if name.startswith('__') and name.endswith('__'):
                continue
            attr = getattr(self, name)
            if attr and not isinstance(attr, (Printable, list, set, tuple, dict, int, str, float, bool)):
                continue
            if not count:
                indent("", level)
            else:
                indent(",", level)
            stream.write(name)
            stream.write(": ")
            write_value(attr, level)
            count += 1
        stream.write(">")
        return stream


@dataclass(frozen=True)
class FilePos(Printable):
    line: int
    token: str


@dataclass
class Type(Printable):
    pos: FilePos
    pred: "Type"|None

    def is_atomic(self) -> bool:
        raise NotImplementedError(f"{type(self)}.is_atomic() is not implemented")


@dataclass
class Nil(Type):
    def is_atomic(self) -> bool:
        return True


@dataclass
class Int(Type):
    def is_atomic(self) -> bool:
        return True


@dataclass
class Str(Type):
    def is_atomic(self) -> bool:
        return True


@dataclass
class Func(Type):
    params: tuple[Type, ...]
    result: Type

    def is_atomic(self) -> bool:
        return False


@dataclass
class Union(Type):
    types: tuple[Type, ...]

    def is_atomic(self) -> bool:
        return False

@dataclass
class Ref(Type):
    _id: int
    value: Type|None

    def is_atomic(self) -> bool:
        return False


def deref(t: Type) -> Type:
    while isinstance(t, Ref) and t.value:
        t = replace(t.value, pred=t)
    return t


@dataclass
class Error(Printable):
    kind: str
    left: Type
    right: Type

    def __repr__(self):
        return Printable.__repr__(self)

    def to_str(self, stream, seen):
        stream.write(f"<Error kind: {self.kind} left: ")
        self.left.to_str(stream, seen)
        stream.write(", right: ")
        self.right.to_str(stream, seen)
        stream.write(">")


class Unifier:
    def __init__(self, ctx, a: Type, b: Type):
        self.ctx = ctx
        self.start = (a, b)
        self.stack: list[tuple[Type, Type]]
        self.errors: list[Error] = []

    def unify(self) -> "Unifier":
        self.stack = [self.start]
        self.errors = []
        while self.stack:
            a, b = self.stack.pop()
            self.unify_step(a, b)
        return self
 
    def bind(self, x: Ref, t: Type) -> None:
        x.value = t

    def copy(self, pos: FilePos) -> "Unifier":
        a, b = self.start
        return Unifier(self.ctx, self.ctx.copy(a), self.ctx.copy(b))
   
    def unify_step(self, a: Type, b: Type) -> None:
        a, b = deref(a), deref(b)
        if isinstance(a, Ref) and not isinstance(b, Ref):
            self.bind(a, b)
            return
        if not isinstance(a, Ref) and isinstance(b, Ref):
            self.bind(b, a)
            return
        if isinstance(a, Ref) and isinstance(b, Ref):
            if a == b:
                return
            # Sort so that 'a' is older than 'b' (lower id) 
            if a._id > b._id:
                a, b = b, a
            # Bind the most recent to the most ancient.
            self.bind(b, a)
            return
        if type(a) != type(b):
            self.error("different", a, b)
            return
        if isinstance(a, Func) and isinstance(b, Func):
            if len(a.params) != len(b.params):
                self.error("func_params_length", a, b)
            # Push params to stack in reverse order, so that they'll be
            # pop'ped in order.
            # If they don't have the same length, assume the last elements
            # are missing from the smallest one.
            n = min(len(a.params), len(b.params))
            for p, q in zip(a.params[n-1::-1], b.params[n-1::-1]):
                p = replace(p, pred=a)
                q = replace(q, pred=b)
                self.stack.append((p, q))
            return
        # Unknown / unhandled cases
        self.error("unknown", a, b)

    def error(self, kind: str, a: Type, b: Type):
        self.errors.append(Error(kind, a, b))


class Checker:
    def __init__(self, text: str):
        self.text = text
        self.ref_id = 0

    def new_ref(self, pos: FilePos, pred=None):
        self.ref_id += 1
        return Ref(pos=pos, pred=pred, _id=self.ref_id, value=None)

    def unify(self, a: Type, b: Type) -> Unifier:
        u = Unifier(self, a, b)
        u.unify()
        return u

    def copy(self, x: Type, pos=None) -> Type:
        pos = pos or x.pos
        if isinstance(x, Ref):
            return self.new_ref(pos, x)
        if isinstance(x, Func):
            params = tuple(self.copy(param) for param in x.params)
            result = self.copy(x.result)
            return Func(pos, x, params, result)
        return replace(x, pos=pos)


def main():
    program = dedent("""
    fun p123(x, f) {
        var a = 123;
        return f(x, a);
    }
    fun add(x, y) {
        return x + y;
    }
    print p123(20000, add);
    print p123("str", add);
    """)

    ctx = Checker(program)

    builtin = FilePos(0, "<builtin>")
    int_, str_ = Int(builtin, None), Str(builtin, None)
    plus = Union(builtin, None, types=(
        Func(builtin, None, (int_, int_), int_),
        Func(builtin, None, (str_, str_), str_),
    ))

    p123_x = ctx.new_ref(FilePos(1, "x"))
    p123_f = ctx.new_ref(FilePos(1, "f"))
    p123_ret = ctx.new_ref(FilePos(1, "p123"))
    p123 = Func(pos=FilePos(1, "p123(x, f)"), pred=None, params=(p123_x, p123_f), result=p123_ret)

    p123_a = ctx.new_ref(FilePos(2, "a"))
    p123_123 = Int(FilePos(2, "123"), None)
    p123_a.value = p123_123

    p123_f_1 = ctx.new_ref(FilePos(3, "x"))
    p123_f_2 = ctx.new_ref(FilePos(3, "a"))
    p123_f_ret = ctx.new_ref(FilePos(3, "f"))
    p123_f_call = Func(FilePos(3, "f(x, a)"), None, params=(p123_f_1, p123_f_2), result=p123_f_ret)
    p123_f_ret.value = p123_ret
    p123_unifier = ctx.unify(p123_f_call, p123_f)
    print("p123:", p123)
    print("p123 errors:", p123_unifier.errors)

    add_x = ctx.new_ref(FilePos(5, "x"))
    add_y = ctx.new_ref(FilePos(5, "y"))
    add_ret = ctx.new_ref(FilePos(5, "add"))
    add = Func(FilePos(5, "add(x, y)"), None, (add_x, add_y), add_ret)

    add_plus_1 = ctx.new_ref(FilePos(6, "x"))
    add_plus_2 = ctx.new_ref(FilePos(6, "y"))
    add_plus_ret = ctx.new_ref(FilePos(6, "+"))
    add_plus_call = Func(FilePos(6, "x + y"), None, (add_plus_1, add_plus_2), add_plus_ret)
    add_plus_ret.value = add_ret
    add_unifier = ctx.unify(add_plus_call, ctx.copy(plus, FilePos(6, "+")))
    print("add:", add)
    print("add errors:", p123_unifier.errors)

    globals_20000 = Int(FilePos(8, "20000"), None)
    globals_add = ctx.copy(add, FilePos(8, "add"))
    print_1_p123 = ctx.copy(p123, FilePos(8, "p123"))
    print_1_p123_1 = globals_20000
    print_1_p123_2 = globals_add
    print_1_p123_ret = ctx.new_ref(FilePos(8, "p123"))
    print_1_p123_call = Func(FilePos(8, "p123(20000, add)"), None, (print_1_p123_1, print_1_p123_2), print_1_p123_ret)
    print_1_p123_unifier = ctx.unify(print_1_p123, print_1_p123_call)
    print("print #1 p123:", print_1_p123)
    print("print #1 errors:", print_1_p123_unifier.errors)

    globals_str = Str(FilePos(9, '"str"'), None)
    globals_add = ctx.copy(add, FilePos(9, "add"))
    print_2_p123 = ctx.copy(p123, FilePos(9, "p123"))
    print_2_p123_1 = globals_str
    print_2_p123_2 = globals_add
    print_2_p123_ret = ctx.new_ref(FilePos(9, "p123"))
    print_2_p123_call = Func(FilePos(9, 'p123("str", add)'), None, (print_2_p123_1, print_2_p123_2), print_2_p123_ret)
    print_2_p123_unifier = ctx.unify(print_2_p123_call, print_2_p123)
    print("print #2 p123:", print_2_p123)
    print("print #2 errors:", print_2_p123_unifier.errors)


if __name__ == '__main__':
    main()
