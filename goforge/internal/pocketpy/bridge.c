#include "pocketpy.h"

extern _Bool goDispatch(char* name, int argc, py_TValue* argv);

static bool bridgeFunc(int argc, py_StackRef argv) {
	py_StackRef func = py_inspect_currentfunction();
	if (func == NULL) {
		py_exception(tp_RuntimeError, "cannot inspect current function");
		return false;
	}
	if (!py_getattr(func, py_name("__name__"))) {
		return false;
	}
	char* funcName = (char*)py_tostr(py_retval());
	return goDispatch(funcName, argc, argv);
}

py_CFunction pk_bridge_ptr = NULL;

__attribute__((constructor)) static void initBridge(void) {
	pk_bridge_ptr = bridgeFunc;
}
