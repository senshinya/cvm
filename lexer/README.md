# lexer

Support all C99 lexical grammar except:

- universal character names
- wchar_t literal (include string and char literal) for extended character set
- [alternative punctuators](https://en.cppreference.com/w/cpp/language/operator_alternative): <:  :>  <%  %>  %:  %:%:
- float constant with leading . (like .5)

and limit the source character set to ASCII