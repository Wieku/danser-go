#version 330

uniform sampler2DArray tex;

in vec2 uv;
in vec4 color;

out vec4 out_color;

void main() {
    out_color = texture(tex, vec3(uv, 0)) * color;//vec4(color.rgb, color.a * texture(tex, vec3(uv, 0)).r);
}