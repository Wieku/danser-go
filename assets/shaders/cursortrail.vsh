#version 330

in vec3 in_position;
in vec3 in_mid;
in vec2 in_tex_coord;
in float in_index;

uniform mat4 proj;
uniform float scale;
uniform float points;
uniform float endScale;

out vec2 tex_coord;
out float index;

void main() {
    gl_Position = proj * vec4((in_position - in_mid) * scale * (endScale + (1.0 - endScale) * (points-1.0-in_index) / points) + in_mid, 1.0);
    tex_coord = in_tex_coord;
	index = in_index;
}