# Lexer (Lexical Analysis)

Support all C99 lexical grammar except:

- universal character names
- wchar_t literal (include string and char literal) for extended character set
- [alternative punctuators](https://en.cppreference.com/w/cpp/language/operator_alternative): <:  :>  <%  %>  %:  %:%:

and limit the source character set to ASCII

Lexical grammar from _ISO/IEC 9899:TC3 A.1 Lexical grammar_