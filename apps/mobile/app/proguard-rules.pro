# Apollo-generated models are reached by reflection-free codegen; no extra rules needed.

# ONNX Runtime's native code calls back into its Java classes by name.
-keep class ai.onnxruntime.** { *; }
