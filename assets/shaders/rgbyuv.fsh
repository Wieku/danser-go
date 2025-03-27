#version 330
precision highp float;

uniform sampler2DArray tex;

in vec2 tex_coord;

layout(location = 0) out float outY;
layout(location = 1) out float outU;
layout(location = 2) out float outV;

// BT.709 matrix according to ITU document: https://www.itu.int/dms_pubrec/itu-r/rec/bt/R-REC-BT.709-6-201506-I!!PDF-E.pdf
// BT.601 matrix according to ITU document: https://www.itu.int/dms_pubrec/itu-r/rec/bt/R-REC-BT.601-7-201103-I!!PDF-E.pdf

const mat4x3 rgbToYuvBT709Full = mat4x3(0.213434, -0.115021,  0.501961,
                                        0.718005, -0.386939, -0.455933,
                                        0.072483,  0.501961, -0.046027,
                                               0,  0.501961,  0.501961);

const mat4x3 rgbToYuvBT601Full = mat4x3(0.300173, -0.169398,  0.501961,
                                        0.589302, -0.332563, -0.420330,
                                        0.114447,  0.501961, -0.081631,
                                               0,  0.501961,  0.501961);

const mat4x3 rgbToYuvBT709TV = mat4x3(0.183302, -0.101039,  0.440938,
                                      0.616640, -0.339900, -0.400505,
                                      0.062250,  0.440938, -0.040431,
                                      0.062745,  0.501961,  0.501961);

const mat4x3 rgbToYuvBT601TV = mat4x3(0.257796, -0.148804,  0.440937,
                                      0.506106, -0.292133, -0.369231,
                                      0.098290,  0.440937, -0.071706,
                                      0.062745,  0.501961,  0.501961);

void main() {
    vec3 src = texture(tex, vec3(tex_coord.x, 1 - tex_coord.y, 0)).rgb;
    vec3 color = rgbToYuvBT601TV*vec4(src, 1);
    outY = color.r;
    outU = color.g;
    outV = color.b;
}
