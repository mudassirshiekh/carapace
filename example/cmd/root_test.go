package cmd

import (
	"testing"

	"github.com/rsteube/carapace"
	"github.com/rsteube/carapace/assert"
)

func TestBash(t *testing.T) {
	expected := `#!/bin/bash
_example_callback() {
  local compline="${COMP_LINE:0:${COMP_POINT}}"
  local last="${COMP_WORDS[${COMP_CWORD}]}"
  if [[ $last =~ ^[\"\'] ]] && ! echo "$last" | xargs echo 2>/dev/null >/dev/null ; then
      compline="${compline}${last:0:1}"
      last="${last/ /\\\\ }" 
  fi

  echo "$compline" | sed -e 's/ $/ _/' -e 's/"/\"/g' | xargs example _carapace bash "$1"
}

_example_completions() {
  local compline="${COMP_LINE:0:${COMP_POINT}}"
  local last="${COMP_WORDS[${COMP_CWORD}]}"
  
  if [[ $last =~ ^[\"\'] ]] && ! echo "$last" | xargs echo 2>/dev/null >/dev/null ; then
      compline="${compline}${last:0:1}"
      last="${last/ /\\\\ }" 
  else
      last="${last/ /\\\ }" 
  fi

  local state=$(echo "$compline" | sed -e "s/ \$/ _/" -e 's/"/\"/g' | xargs example _carapace bash state)
  local previous="${COMP_WORDS[$((${COMP_CWORD}-1))]}"
  local IFS=$'\n'

  case $state in

    '_example' )
      if [[ $last == -* ]]; then
        COMPREPLY=($(compgen -W $'--array\n-a\n--persistentFlag\n-p\n--toggle\n-t' -- $last))
      else
        case $previous in
          -a | --array)
            COMPREPLY=($())
            ;;



          *)
            COMPREPLY=($(compgen -W $'action\nalias\ncallback\ncondition\ninjection' -- $last))
            ;;
        esac
      fi
      ;;


    '_example__action' )
      if [[ $last == -* ]]; then
        COMPREPLY=($(compgen -W $'--custom\n-c\n--directories\n--files\n-f\n--groups\n-g\n--hosts\n--message\n-m\n--multi_parts\n--net_interfaces\n-n\n--users\n-u\n--values\n-v\n--values_described\n-d' -- $last))
      else
        case $previous in
          -c | --custom)
            COMPREPLY=($())
            ;;

          --directories)
            COMPREPLY=($(compgen -S / -d -- $last))
            ;;

          -f | --files)
            COMPREPLY=($(compgen -S / -d -- $last; compgen -f -X '!*.go' -- $last))
            ;;

          -g | --groups)
            COMPREPLY=($(compgen -g -- $last))
            ;;

          --hosts)
            COMPREPLY=($(compgen -W "$(cat ~/.ssh/known_hosts | cut -d ' ' -f1 | cut -d ',' -f1)" -- $last))
            ;;

          -m | --message)
            COMPREPLY=($(compgen -W $'ERR\nmessage\\\ example' -- $last))
            ;;

          --multi_parts)
            COMPREPLY=($(compgen -W $'multi/parts\nmulti/parts/example\nmulti/parts/test\nexample/parts' -- $last))
            ;;

          -n | --net_interfaces)
            COMPREPLY=($(compgen -W "$(ifconfig -a | grep -o '^[^ :]\+')" -- $last))
            ;;

          -u | --users)
            COMPREPLY=($(compgen -u -- $last))
            ;;

          -v | --values)
            COMPREPLY=($(compgen -W $'values\nexample' -- $last))
            ;;

          -d | --values_described)
            COMPREPLY=($(compgen -W $'values\nexample\n\n' -- $last))
            ;;

          *)
            COMPREPLY=($(eval $(_example_callback '_')))
            ;;
        esac
      fi
      ;;


    '_example__callback' )
      if [[ $last == -* ]]; then
        COMPREPLY=($(compgen -W $'--callback\n-c' -- $last))
      else
        case $previous in
          -c | --callback)
            COMPREPLY=($(eval $(_example_callback '_example__callback##callback')))
            ;;

          *)
            COMPREPLY=($(eval $(_example_callback '_')))
            ;;
        esac
      fi
      ;;


    '_example__condition' )
      if [[ $last == -* ]]; then
        COMPREPLY=($(compgen -W $'--required\n-r' -- $last))
      else
        case $previous in
          -r | --required)
            COMPREPLY=($(compgen -W $'valid\ninvalid' -- $last))
            ;;

          *)
            COMPREPLY=($(eval $(_example_callback '_')))
            ;;
        esac
      fi
      ;;


    '_example__injection' )
      if [[ $last == -* ]]; then
        COMPREPLY=($())
      else
        case $previous in

          *)
            COMPREPLY=($(eval $(_example_callback '_')))
            ;;
        esac
      fi
      ;;

  esac

  [[ $last =~ ^[\"\'] ]] && COMPREPLY=("${COMPREPLY[@]/\\ /\ }")
  [[ $COMPREPLY == */ ]] && compopt -o nospace
}

complete -F _example_completions example
`
	assert.Equal(t, expected, carapace.Gen(rootCmd).Bash())
}

func TestFish(t *testing.T) {
	expected := `function _example_state
  set -lx CURRENT (commandline -cp)
  if [ "$LINE" != "$CURRENT" ]
    set -gx LINE (commandline -cp)
    set -gx STATE (commandline -cp | xargs example _carapace fish state)
  end

  [ "$STATE" = "$argv" ]
end

function _example_callback
  set -lx CALLBACK (commandline -cp | sed "s/ \$/ _/" | xargs example _carapace fish $argv )
  eval "$CALLBACK"
end

complete -c example -f

complete -c example -f -n '_example_state _example' -l array -s a -d 'multiflag' -r
complete -c example -f -n '_example_state _example' -l persistentFlag -s p -d 'Help message for persistentFlag'
complete -c example -f -n '_example_state _example' -l toggle -s t -d 'Help message for toggle' -a '(echo -e "true\nfalse")' -r
complete -c example -f -n '_example_state _example ' -a 'action alias' -d 'action example'
complete -c example -f -n '_example_state _example ' -a 'callback ' -d 'callback example'
complete -c example -f -n '_example_state _example ' -a 'condition ' -d 'condition example'
complete -c example -f -n '_example_state _example ' -a 'injection ' -d 'just trying to break things'


complete -c example -f -n '_example_state _example__action' -l custom -s c -d 'custom flag' -a '()' -r
complete -c example -f -n '_example_state _example__action' -l directories -d 'files flag' -a '(__fish_complete_directories)' -r
complete -c example -f -n '_example_state _example__action' -l files -s f -d 'files flag' -a '(__fish_complete_suffix ".go")' -r
complete -c example -f -n '_example_state _example__action' -l groups -s g -d 'groups flag' -a '(__fish_complete_groups)' -r
complete -c example -f -n '_example_state _example__action' -l hosts -d 'hosts flag' -a '(__fish_print_hostnames)' -r
complete -c example -f -n '_example_state _example__action' -l message -s m -d 'message flag' -a '(echo -e "ERR\tmessage example\n_")' -r
complete -c example -f -n '_example_state _example__action' -l multi_parts -d 'multi_parts flag' -a '(echo -e "multi/parts\nmulti/parts/example\nmulti/parts/test\nexample/parts")' -r
complete -c example -f -n '_example_state _example__action' -l net_interfaces -s n -d 'net_interfaces flag' -a '(__fish_print_interfaces)' -r
complete -c example -f -n '_example_state _example__action' -l users -s u -d 'users flag' -a '(__fish_complete_users)' -r
complete -c example -f -n '_example_state _example__action' -l values -s v -d 'values flag' -a '(echo -e "values\nexample")' -r
complete -c example -f -n '_example_state _example__action' -l values_described -s d -d 'values with description flag' -a '(echo -e "values\tvalueDescription\nexample\texampleDescription\n\n")' -r
complete -c example -f -n '_example_state _example__action' -a '(_example_callback _)'


complete -c example -f -n '_example_state _example__callback' -l callback -s c -d 'Help message for callback' -a '(_example_callback _example__callback##callback)' -r
complete -c example -f -n '_example_state _example__callback' -a '(_example_callback _)'


complete -c example -f -n '_example_state _example__condition' -l required -s r -d 'required flag' -a '(echo -e "valid\ninvalid")' -r
complete -c example -f -n '_example_state _example__condition' -a '(_example_callback _)'


complete -c example -f -n '_example_state _example__injection' -a '(_example_callback _)'
`
	assert.Equal(t, expected, carapace.Gen(rootCmd).Fish())
}

func TestZsh(t *testing.T) {
	expected := `#compdef example
function _example {
  local -a commands
  # shellcheck disable=SC2206
  local -a -x os_args=(${words})

  _arguments -C \
    "(*-a *--array)"{\*-a,\*--array}"[multiflag]: :" \
    "(-p --persistentFlag)"{-p,--persistentFlag}"[Help message for persistentFlag]" \
    "(-t --toggle)"{-t,--toggle}"[Help message for toggle]: :_values '' true false" \
    "1: :->cmnds" \
    "*::arg:->args"

  # shellcheck disable=SC2154
  case $state in
    cmnds)
      # shellcheck disable=SC2034
      commands=(
        "action:action example"
        "alias:action example"
        "callback:callback example"
        "condition:condition example"
        "injection:just trying to break things"
      )
      _describe "command" commands
      ;;
  esac
  
  case "${words[1]}" in
    action)
      _example__action
      ;;
    alias)
      _example__action
      ;;
    callback)
      _example__callback
      ;;
    condition)
      _example__condition
      ;;
    injection)
      _example__injection
      ;;
  esac
}

function _example__action {
    _arguments -C \
    "(-c --custom)"{-c,--custom}"[custom flag]: :_most_recent_file 2" \
    "--directories[files flag]: :_files -/" \
    "(-f --files)"{-f,--files}"[files flag]: :_files -g '*.go'" \
    "(-g --groups)"{-g,--groups}"[groups flag]: :_groups" \
    "--hosts[hosts flag]: :_hosts" \
    "(-m --message)"{-m,--message}"[message flag]: : _message -r 'message example'" \
    "--multi_parts[multi_parts flag]: :_multi_parts / '(multi/parts multi/parts/example multi/parts/test example/parts)'" \
    "(-n --net_interfaces)"{-n,--net_interfaces}"[net_interfaces flag]: :_net_interfaces" \
    "(-u --users)"{-u,--users}"[users flag]: :_users" \
    "(-v --values)"{-v,--values}"[values flag]: :_values '' values example" \
    "(-d --values_described)"{-d,--values_described}"[values with description flag]: :_values '' 'values[valueDescription]' 'example[exampleDescription]'  " \
    "1:: :_values '' positional1 p1" \
    "2:: :_values '' positional2 p2"
}

function _example__callback {
    _arguments -C \
    "(-c --callback)"{-c,--callback}"[Help message for callback]: : eval \$(${os_args[1]} _carapace zsh '_example__callback##callback' ${${os_args:1:gs/\"/\\\"}:gs/\'/\\\"})" \
    "1:: : eval \$(${os_args[1]} _carapace zsh '_example__callback#1' ${${os_args:1:gs/\"/\\\"}:gs/\'/\\\"})"
}

function _example__condition {
    _arguments -C \
    "(-r --required)"{-r,--required}"[required flag]: :_values '' valid invalid" \
    "1:: : eval \$(${os_args[1]} _carapace zsh '_example__condition#1' ${${os_args:1:gs/\"/\\\"}:gs/\'/\\\"})"
}

function _example__injection {
    _arguments -C \
    "1:: :_values '' echo\ fail" \
    "2:: :_values '' echo\ fail" \
    "3:: :_values '' echo\ fail" \
    "4:: :_values '' \ echo\ fail\ " \
    "5:: :_values '' \ echo\ fail\ " \
    "6:: :_values '' \ echo\ fail\ " \
    "7:: :_values '' echo\ fail" \
    "8:: : _message -r 'no values to complete'" \
    "9:: :_values '' LAST\ POSITIONAL\ VALUE"
}
if compquote '' 2>/dev/null; then _example; else compdef _example example; fi
`
	assert.Equal(t, expected, carapace.Gen(rootCmd).Zsh())
}
