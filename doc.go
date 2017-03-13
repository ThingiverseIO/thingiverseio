/*Package thingiverseio provides a golang implementation of ThingiverseIO. 
 
ThingiverseIO is peer to peer remote procedure call protocoll with focus on ease of use. It is designed to enable beginner level programmers to build small scale distributed systems. It was specifically developed to meet the requirements of modern scientific laboratories.

There 2 types of ThingiverseIO nodes: Inputs and Outputs. Whereas Inputs import functionality from the network, Outputs export functionality. Peer discovery is done via function descriptors of the form

	function FUNCTION_NAME(INPUTPARAMETER_NAME INPUTPARAMETER_TYPE, ...)(OUTPUTPARAMETER_NAME OUTPUTPARAMETER_TYPE, ...)
	tags TAG, ...

where names can be freely chosen by the user and types must be either "bin", "string", "int", "float" or "bool". Arrays are denoted by prepending one "[]" for every dimension to the type.

Inputs will connect to every Output which exports all function signatures and tags described in the Inputs function descriptor. The Output may export more functions/tags then required by the Input.

Tags can either be a single string or a key value pair. A line containg a single tag starts with "tag", a multitag line with "tags". Tags are separeted with space, key value tags are noted as "KEY:VALUE". Example:

	tag simple_tag
	tag key_tag:tag_value
	tags multisimple muiltikey:val

*/
package thingiverseio

