#version 330

uniform sampler2DArray tex;

uniform int texRGBA;

in vec2 uv;
in vec4 color;

out vec4 out_color;

void main() {
    vec4 cColor = texRGBA == 0 ? vec4(vec3(1.0), texture(tex, vec3(uv, 0)).r) : texture(tex, vec3(uv, 0));

    out_color = cColor * color;//vec4(color.rgb, color.a * texture(tex, vec3(uv, 0)).r);
}