#version 330
precision highp float;

uniform sampler2DArray tex;

in vec2 tex_coord;

out vec4 color;

// BT.709 matrix according to ITU document: https://www.itu.int/dms_pubrec/itu-r/rec/bt/R-REC-BT.709-6-201506-I!!PDF-E.pdf
const mat3 rgbToYuv = mat3(0.2126, -0.11457210605,            0.5,  // [        0.2126,         0.7152,         0.0722]
                           0.7152, -0.38542789394,  -0.4541529083,  // [-0.11457210605, -0.38542789394,            0.5]
                           0.0722,            0.5, -0.04584709169); // [           0.5,  -0.4541529083, -0.04584709169]

const vec3 correction = vec3(0.0625, 0.5, 0.5); //[16, 128, 128]

void main() {
    vec3 src = texture(tex, vec3(tex_coord, 0)).rgb;

    color = vec4(rgbToYuv*src+correction, 0);
}
