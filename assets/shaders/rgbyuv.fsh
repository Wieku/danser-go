#version 330
precision highp float;

uniform sampler2DArray tex;

in vec2 tex_coord;

out vec4 color;

// BT.709 matrix according to ITU document: https://www.itu.int/dms_pubrec/itu-r/rec/bt/R-REC-BT.709-6-201506-I!!PDF-E.pdf
// BT.601 matrix according to ITU document: https://www.itu.int/dms_pubrec/itu-r/rec/bt/R-REC-BT.601-7-201103-I!!PDF-E.pdf

const mat4x3 rgbToYuvBT709Full = mat4x3(0.2126, -0.114572,       0.5,
                                        0.7152, -0.385428, -0.454152,
                                        0.0722,       0.5, -0.045847,
                                             0,       0.5,       0.5);

const mat4x3 rgbToYuvBT601Full = mat4x3(0.299, -0.168736,       0.5,
                                        0.587, -0.331264, -0.418688,
                                        0.114,       0.5, -0.081312,
                                            0,       0.5,       0.5);

const mat4x3 rgbToYuvBT709TV = mat4x3(0.182586, -0.100644,  0.439216,
                                      0.614231, -0.338572, -0.398941,
                                      0.062007,  0.439216, -0.040273,
                                        0.0625,       0.5,       0.5);

const mat4x3 rgbToYuvBT601TV = mat4x3(0.256788, -0.148223,  0.439216,
                                      0.504129, -0.290993, -0.367789,
                                      0.097906,  0.439216, -0.071427,
                                        0.0625,       0.5,       0.5);

void main() {
    vec3 src = texture(tex, vec3(tex_coord, 0)).rgb;
    color = vec4(rgbToYuvBT601TV*vec4(src, 1), 0);
}
