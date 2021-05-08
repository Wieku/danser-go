#version 330

in vec4 col_tint;
in vec3 tex_coord;
in float additive;

uniform sampler2DArray tex;

out vec4 color;

float median(float r, float g, float b) {
    return max(min(r, g), min(max(r, g), b));
}

void main() {
    vec4 in_color = texture(tex, tex_coord);

    color = in_color*col_tint;
    color.rgb *= color.a;
    color.a *= additive;
}