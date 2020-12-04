#version 330

in vec2 in_position;
in vec2 in_tex_coord;

in vec2 in_mid;
in float fadeTime;

uniform mat4 proj;
uniform float scale;
uniform float clock;

uniform vec2 u;
uniform vec2 v;

out vec2 tex_coord;
out vec4 color_pass;

void main() {
    gl_Position = proj * vec4(in_position * scale + in_mid, 0.0, 1.0);
    tex_coord = vec2(u.x, v.x) + in_tex_coord * vec2(u.y-u.x, v.y-v.x);

    color_pass = vec4(vec3(1.0), clamp((fadeTime-clock)/3, 0.0, 1.0));
}