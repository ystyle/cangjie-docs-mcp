# 仓颉语言基础语法
>仓颉语言(Cangjie lang), 文件后缀为cj, 简称为cj

### 变量
修饰符 变量名: 变量类型 = 初始值

其中修饰符用于设置变量的各类属性，可以有一个或多个，常用的修饰符包括：

可变性修饰符：let 与 var，分别对应不可变和可变属性，可变性决定了变量被初始化后其值还能否改变，仓颉变量也由此分为不可变变量和可变变量两类(注册let不能像rust那种重新定义一个同名的变量shawdow)。

可见性修饰符：private 与 public 等，影响全局变量和成员变量的可引用范围，详见后续章节的相关介绍。
静态性修饰符：static，影响成员变量的存储和引用方式，详见后续章节的相关介绍。
变量均支持赋值操作符（=），与类型无关。let 修饰的变量只能被赋值一次，即初始化，var 修饰的变量可以被多次赋值。

```cj
main() {
    let a: Int64 = 20
    var b: Int64 = 12
    b = 23
    println("${a}${b}")
}
```
### 基础类型
- 有符号整数类型包括 Int8、Int16、Int32、Int64 和 IntNative
- 无符号整数类型包括 UInt8、UInt16、UInt32、UInt64 和 UIntNative
- 浮点类型包括 Float16、 Float32 和 Float64
- 布尔类型只有两个字面量：true 和 false。
- 字符类型使用 Rune 表示，可以表示 Unicode 字符集中的所有字符。`let a: Rune = r'a'`
  - Rune 到 UInt32 的转换使用 UInt32(e) 的方式，其中 e 是一个 Rune 类型的表达式，UInt32(e) 的结果是 e 的 Unicode scalar value 对应的 UInt32 类型的整数值。
  - 整数类型到 Rune 的转换使用 Rune(num) 的方式，其中 num 的类型可以是任意的整数类型，且仅当 num 的值落在 [0x0000, 0xD7FF] 或 [0xE000, 0x10FFFF] （即 Unicode scalar value）中时，返回对应的 Unicode scalar value 表示的字符，否则，编译报错（编译时可确定 num 的值）或运行时抛异常。
  - 转义字符`let slash: Rune = r'\\'`, `let newLine: Rune = r'\n'`, `let tab: Rune = r'\t'`
  - 仓颉源代码可以直接用Unicode字面量定义字符， 通用字符以 \u 开头，后面加上定义在一对花括号中的 1~8 个十六进制数，即可表示对应的 Unicode 值代表的字符。举例如下：
    - `let he: Rune = r'\u{4f60}'` // 你
    - `let llo: Rune = r'\u{597d}'` // 好
- 字符串类型使用 String 表示，用于表达文本数据，由一串 Unicode 字符组合而成。`let s2 = 'Hello Cangjie Lang'`
  - 插值字符串
  ```cj
  let fruit = "apples"
  let count = 10
  let s = "There are ${count * count} ${fruit}"
  ```
  - 字符串转为Rune数组： `"apples".toRuneArray()`
  - 多行原始字符串字面量以一个或多个井号（#）和一个单引号（'）或双引号（"）开头，后跟任意数量的合法字符，直到出现与字符串开头相同的引号和与字符串开头相同数量的井号为止。在当前文件结束之前，如果还没遇到匹配的双引号和相同个数的井号，则编译报错。与多行字符串字面量一样，原始多行字符串字面量可以跨越多行。不同之处在于，转义规则不适用于多行原始字符串字面量，字面量中的内容会维持原样（转义字符不会被转义，如下例中 s2 中的 \n 不是换行符，而是由 \ 和 n 组成的字符串 \n）。
  ```cj
  let s1: String = #""#
  let s2 = ##'\n'##
  let s3 = ###"
    Hello,
    Cangjie
    Lang"###
  ```
- 元组（Tuple）可以将多个不同的类型组合在一起，成为一个新的类型。元组类型使用 (T1, T2, ..., TN) 表示，其中 T1 到 TN 可以是任意类型，不同类型间使用逗号（,）连接。元组至少是二元，例如，(Int64, Float64) 表示一个二元组类型，(Int64, Float64, String) 表示一个三元组类型。`var tuple = (true, false)` ` println(tuple[0])`
- 数组类型:仓颉使用 Array<T> 来表示 Array 类型。T 表示 Array 的元素类型，T 可以是任意类型。
- 区间类型用于表示拥有固定步长的序列，区间类型是一个泛型（详见泛型章节），使用 Range<T> 表示。
  - 每个区间类型的实例都会包含 start、end 和 step 三个值。其中，start 和 end 分别表示序列的起始值和终止值，step 表示序列中前后两个元素之间的差值（即步长）；start 和 end 的类型相同（即 T 被实例化的类型），step 类型是 Int64，并且它的值不能等于 0。
- Unit 类型只有一个值，也是它的字面量：()。除了赋值、判等和判不等外，Unit 类型不支持其他操作。

### 表达式
>表达式`if`,`for-in`, `while`, `go-while`条件的括号不能省略
- `if`表达式的基本形式：
```cj
if (条件) {
  分支 1
} else {
  分支 2
}
```
一个“let pattern”的构成为 let pattern <- expression，其中各字段含义为：

pattern ：模式，用于匹配 expression 的值类型和内容。
<- ：模式匹配操作符。
expression ：表达式，该表达式求值后，再和模式进行匹配。expression 表达式的优先级不能低于 .. 运算符，但是可以用 () 改变优先级。运算符优先级请参见操作符。
当表达式和模式匹配成功时，该模式匹配的值为 true，此时执行 if 分支对应的代码块；反之，为 false，执行 else 分支代码块，else 分支可以不存在。
```cj
let a = Some(3)
let d = Some(1)
if (let Some(e) <- a && let Some(f) <- d) { // both patterns match, value of condition is true
   println("${e} ${f}") // print 3 1
}
```
- while 表达式
```cj
while (条件) {
  循环体
}
do {
  循环体
} while (条件)
```
- for-in 表达式
```cj
for (迭代变量 in 序列) {
  循环体
}
let array = [(1, 2), (3, 4), (5, 6)]
for ((x, y) in array) {
    println("${x}, ${y}")
}
// 遍历区间类型实例
main() {
    var sum = 0
    for (i in 1..=100) {
        sum += i
    }
    println(sum)
}
```
几个循环都可以使用break,continue, break和continue都不支持标签跳转，且cj里完全不支持goto


### 函数
仓颉编程语言中，函数是一等公民（first-class citizens），可以作为函数的参数或返回值，也可以赋值给变量。因此函数本身也有类型，称之为函数类型。

函数类型由函数的参数类型和返回类型组成，参数类型和返回类型之间使用 -> 连接。参数类型使用圆括号 () 括起来，可以有 0 个或多个参数，如果参数超过一个，参数类型之间使用逗号（,）分隔。
函数参数也是默认是let定义的不可变变量

```cj
// 如a和b都是不可变的
func add(a: Int64, b: Int64): Int64 {
    return a + b
}

type FnType = (Int64) -> Unit
// display 实现了FnType
func display(a: Int64): Unit {
    println(a)
}
// 返回类型
func add(a: Int64, b: Int64): Int64 {
    a + b
}

func returnAdd(): (Int64, Int64) -> Int64 {
    add
}
```
#### Lambda 表达式定义
Lambda 表达式的语法为如下形式： { p1: T1, ..., pn: Tn => expressions | declarations }。
```cj
let f1 = { a: Int64, b: Int64 => a + b }

var display = { =>   // Parameterless lambda expression.
    println("Hello")
    println("World")
}

// The parameter types are inferred from the type of the variable sum1
var sum1: (Int64, Int64) -> Int64 = { a, b => a + b }

var sum2: (Int64, Int64) -> Int64 = { a: Int64, b => a + b }

func f(a1: (Int64) -> Int64): Int64 {
    a1(1)
}

main(): Int64 {
    // The parameter type of lambda is inferred from the type of function f
    f({ a2 => a2 + 10 })
}

Lambda 表达式支持立即调用
let r2 = { => 123 }()                          // r2 = 123
var g = { x: Int64 => println("x = ${x}") }
g(2)
```

### 枚举类型
定义 enum 时需要把它所有可能的取值一一列出，称这些值为 enum 的构造器（或者 constructor）。

enum 类型的定义以关键字 enum 开头，接着是 enum 的名字，之后是定义在一对花括号中的 enum 体，enum 体中定义了若干构造器，多个构造器之间使用 | 进行分隔（第一个构造器之前的 | 是可选的）。构造器可以是有名字的，也可以是没有名字的 ...。

每个 enum 中至少存在一个有名字的构造器。有名字的构造器可以没有参数（即“无参构造器”），也可以携带若干个参数（即“有参构造器”）。如下示例代码定义了一个名为 RGBColor 的 enum 类型，它有 3 个构造器：Red、Green 和 Blue，分别表示 RGB 色彩模式中的红色、绿色和蓝色。每个构造器有一个 UInt8 类型的参数，用来表示每个颜色的亮度级别。
```cj
enum RGBColor {
    | Red(UInt8) | Green(UInt8) | Blue(UInt8)
}
```
枚举的模式匹配支持：常量模式、通配符模式、绑定模式、tuple 模式、类型模式和 enum 模式。

# match 表达式的定义:
```cj
main() {
    let x = 0
    match (x) {
        case 1 => let r1 = "x = 1"
                  print(r1)
        case 0 => let r2 = "x = 0" // Matched.
                  print(r2)
        case 2 | 3 | 4 => print("other:") 
        case _ => let r3 = "x != 1 and x != 0"
                  print(r3)
    }
}
```
常量模式: 可以是整数字面量、浮点数字面量、字符字面量、布尔字面量、字符串字面量（不支持字符串插值）、Unit 字面量。
通配符模式: 使用下划线 _ 表示，可以匹配任意值。通配符模式通常作为最后一个 case 中的模式，用来匹配其他 case 未覆盖到的情况
绑定模式: 使用 id 表示，id 是一个合法的标识符。与通配符模式相比，绑定模式同样可以匹配任意值，但绑定模式会将匹配到的值与 id 进行绑定，在 => 之后可以通过 id 访问其绑定的值。
   ```cj
    main() {
        let x = -10
        let y = match (x) {
            case 0 => "zero"
            case n => "x is not zero and x = ${n}" // Matched.
        }
        println(y)
    }
    ```
tuple模式: 用于 tuple 值的匹配，它的定义和 tuple 字面量类似：(p_1, p_2, ..., p_n)
   ```cj
    main() {
        let tv = ("Alice", 24)
        let s = match (tv) {
            case ("Bob", age) => "Bob is ${age} years old"
            case ("Alice", age) => "Alice is ${age} years old" // Matched, "Alice" is a constant pattern, and 'age' is a variable pattern.
            case (name, 100) => "${name} is 100 years old"
            case (_, _) => "someone"
        }
        println(s)
    }
    ```
类型模式: 用于判断一个值的运行时类型是否是某个类型的子类型。
    ```cj
    main() {
        var d = Derived()
        var r = match (d) {
            case b: Base => b.a // Matched.
            case _ => 0
        }
        println("r = ${r}")
    }
    ```
enum模式: 用于匹配 enum 类型的实例，它的定义和 enum 的构造器类似：无参构造器 C 或有参构造器 C(p_1, p_2, ..., p_n)，构造器的类型前缀可以省略，区别在于这里的 p_1 到 p_n（n 大于等于 1）是模式。
    ```cj
    enum TimeUnit {
        | Year(UInt64)
        | Month(UInt64)
    }

    main() {
        let x = Year(2)
        let s = match (x) {
            case Year(n) => "x has ${n * 12} months" // Matched.
            case TimeUnit.Month(n) => "x has ${n} months"
        }
        println(s)
    }
    ```
Tuple 模式和 enum 模式可以嵌套任意模式
```cj
enum TimeUnit {
    | Year(UInt64)
    | Month(UInt64)
}

enum Command {
    | SetTimeUnit(TimeUnit)
    | GetTimeUnit
    | Quit
}

main() {
    let command = (SetTimeUnit(Year(2022)), SetTimeUnit(Year(2024)))
    match (command) {
        case (SetTimeUnit(Year(year)), _) => println("Set year ${year}")
        case (_, SetTimeUnit(Month(month))) => println("Set month ${month}")
        case _ => ()
    }
}
```



### Option类型
Option 类型使用 enum 定义，它包含两个构造器：Some 和 None。其中，Some 会携带一个参数，表示有值，None 不带参数，表示无值。当需要表示某个类型可能有值，也可能没有值的时候，可选择使用 Option 类型。
```cj
enum Option<T> {
    | Some(T)
    | None
}
```
### 类型
class 类型的定义以关键字 class 开头，后跟 class 的名字，接着是定义在一对花括号中的 class 定义体。class 定义体中可以定义一系列的成员变量、成员属性（参见属性）、静态初始化器、构造函数、成员函数和操作符函数
```cj
class Rectangle {
    let width: Int64
    let height: Int64

    public init(width: Int64, height: Int64) {
        this.width = width
        this.height = height
    }

    public func area() {
        width * height
    }
}
let rec = Rectangle(10, 20)
let l = rec.height // l = 20
```
#### class 成员的访问修饰符
对于 class 的成员（包括成员变量、成员属性、构造函数、成员函数），可以使用的访问修饰符有 4 种访问修饰符修饰：private、internal、protected 和 public，缺省的含义是 internal。

- private 表示在 class 定义内可见。
- internal 表示仅当前包及子包（包括子包的子包，详见包章节）内可见。
- protected 表示当前模块（详见包章节）及当前类的子类可见。
- public 表示模块内外均可见。

### 单元测试
- `@Test` 宏应用于顶级函数或顶级类，使该函数或类转换为单元测试类。
- `@TestCase` 宏用于标记单元测试类内的函数，使这些函数成为单元测试的测试用例。
- `@Assert` 声明 Assert 断言，测试函数内部使用，断言失败停止用例。
  - `@Assert(leftExpr, rightExpr)`
  - `@Assert(condition: Bool)`
- `@Expect` 声明 Expect 断言，测试函数内部使用，断言失败继续执行用例。
  - `@Expect(leftExpr, rightExpr)`
  - `@Expect(condition: Bool)`
```cj
@Test
class LexerTest {
    @TestCase
    func test() {
        let a = 1
        // 方式1
        if (a != 1) {
            @Fail("a is not 1")
        }
        // 方式二
        @Assert( a != 1 )
        // 方式三
        @Assert( a,  1 )
    }
}
```
